// Package format define o formato binário `.hwr` de replay de partidas.
//
// Layout do arquivo (todos os inteiros multi-byte em big-endian, salvo o delta
// das entradas que é varint com zig-zag):
//
//	[4 bytes] comprimento do corpo gzipado (uint32)
//	[corpo gzipado]
//	[64 bytes opcionais] assinatura Ed25519
//
// O corpo (antes de gzipar) é:
//
//	magic "HWR1" | versão do formato (u8) | versão do motor (u16)
//	| kid (u8 len + bytes) | matchId (16) | semente (16) | início Unix ms (i64)
//	| modo (u8 len + bytes) | jogadores (u8 count + índice u8 + nome u8 len + bytes)
//	| flags (u8)
//	| entradas (u16 count + ...) | checkpoints (u16 count + 8 bytes cada)
//
// A assinatura cobre `"HWR1-SIG" || corpo gzipado` e fica fora do gzip, no fim
// do arquivo. Arquivo sem assinatura (kid vazio) é aceito como não oficial.
//
// O leitor é hostil: recusa acima de 256 KiB antes de ler, limita o gunzip a
// 1 MiB, 2..10 jogadores, nomes ≤ 32 bytes, 5.000 entradas, payload ≤ 2 KiB,
// delta ≥ 0 e soma ≤ 6h, tipo/índice válidos; todo erro estrutural devolve
// ErrInvalid ("replay inválido") sem ecoar bytes e sem panic.
package format

import (
	"bytes"
	"compress/gzip"
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

// Limites e constantes do formato.
const (
	// Magic é o identificador do arquivo (primeiro campo do corpo descomprimido).
	Magic = "HWR1"

	// FormatVersion é a versão atual do formato .hwr.
	FormatVersion = 1

	// MaxFileSize é o teto do arquivo lido (antes de qualquer leitura do conteúdo).
	MaxFileSize = 256 * 1024
	// MaxDecompressedSize é o teto do corpo após o gunzip (defesa contra bomba de gzip).
	MaxDecompressedSize = 1 * 1024 * 1024

	MinPlayers         = 2
	MaxPlayers         = 10
	MaxPlayerNameBytes = 32
	MaxModeBytes       = 32
	MaxKidBytes        = 32
	// MaxPlayerIDBytes é o teto do id (UUID canônico, 36 bytes) do jogador no cabeçalho.
	MaxPlayerIDBytes = 36
	// MaxRoomIDBytes é o teto do código da sala (room.ID).
	MaxRoomIDBytes = 32
	// MaxRoomNameBytes é o teto do nome da sala.
	MaxRoomNameBytes = 32
	// MaxPasswordBytes é o teto da senha da sala (validate max=24 no gameServer).
	MaxPasswordBytes = 24
	MaxEntries       = 5000
	MaxPayloadBytes  = 2048
	// MaxDurationMs é a duração máxima de uma partida (6 horas).
	MaxDurationMs = 6 * 60 * 60 * 1000

	// CheckpointInterval é a cada quantas entradas um checkpoint é gravado.
	CheckpointInterval = 20
	// CheckpointSize é o tamanho do checkpoint (8 bytes do sha256 do estado).
	CheckpointSize = 8

	// SignatureSize é o tamanho da assinatura Ed25519.
	SignatureSize = ed25519.SignatureSize
	// SignatureDomain é o prefixo de domínio assinado junto com o corpo gzipado.
	SignatureDomain = "HWR1-SIG"
)

// NoPlayer marca entradas que não se referem a um jogador (início/tick).
const NoPlayer = 0xFF

// EntryType é a lista fechada de tipos de entrada.
type EntryType uint8

const (
	EntryStart      EntryType = iota // 0: início
	EntryPlay                        // 1: jogada (payload = JSON original)
	EntryTick                        // 2: tick
	EntryConnect                     // 3: conectou
	EntryDisconnect                  // 4: desconectou
	EntryJoin                        // 5: entrou na sala
	EntryLeave                       // 6: saiu da sala
	EntryEnd                         // 7: fim (Player = índice do vencedor)
)

// entryFlag* marca os campos opcionais de uma entrada.
const (
	entryFlagConnected = 0x01
	entryFlagPayload   = 0x02
)

// Erros tipados. ErrInvalid é o único erro estrutural do leitor, sempre com a
// mensagem fixa "replay inválido" (sem eco de bytes). Os demais indicam que o
// arquivo é estruturalmente válido mas a assinatura não confere ("não oficial").
var (
	ErrInvalid          = errors.New("replay inválido")
	ErrUnknownKey       = errors.New("replay inválido: chave de assinatura desconhecida")
	ErrInvalidSignature = errors.New("replay inválido: assinatura inválida")
)

// Player identifica um jogador na partida. Index é o índice canônico do jogador
// (0..N-1 na ordem lexicográfica dos ids reais, que é a ordem de SortedPlayerIDs
// do motor) e Name é o nome de exibição (até 32 bytes). A ordem da lista Players
// é a ordem de MoveTime, usada pelo runner para reconstruir MoveTime denso.
type Player struct {
	Index uint8
	ID    string
	Name  string
}

// Entry é uma entrada do replay. DeltaMs é a distância em ms desde a entrada
// anterior (ou desde o início na primeira); Connected é o bitmap de conectados
// (nil = sem mudança; senão ceil(nPlayers/8) bytes, bit i = jogador i);
// Payload é o JSON original da jogada (nil = sem payload).
type Entry struct {
	DeltaMs   int64
	Type      EntryType
	Player    uint8
	Connected []byte
	Payload   []byte
}

// Checkpoint são os primeiros 8 bytes do sha256 do estado canônico da sala.
type Checkpoint [CheckpointSize]byte

// Replay é a representação em memória de um replay.
type Replay struct {
	FormatVersion int
	EngineVersion int
	Kid           string
	MatchID       [16]byte
	Seed          [16]byte
	StartedAtMs   int64
	Mode          string
	// RoomID é o código da sala (room.ID), necessário para reconstruir a sala idêntica.
	RoomID      string
	RoomName    string
	Password    string
	MaxPlayers  int
	CustomMatch bool
	CreatedAtMs int64
	Players     []Player
	Flags       uint8

	Entries     []Entry
	Checkpoints []Checkpoint

	// Official e Signature são preenchidos apenas pelo Decode: Official é true
	// quando a assinatura está presente e confere; Signature guarda os 64 bytes
	// (nil quando o arquivo não é assinado).
	Official  bool
	Signature []byte
}

// Signer assina o corpo do replay.
type Signer interface {
	Kid() string
	Sign(message []byte) ([]byte, error)
}

// KeyLookup resolve um kid para a chave pública Ed25519 correspondente.
type KeyLookup func(kid string) (ed25519.PublicKey, bool)

// Ed25519Signer implementa Signer com uma chave privada Ed25519 e um kid.
type Ed25519Signer struct {
	kid string
	key ed25519.PrivateKey
}

// NewEd25519Signer cria um Signer a partir de um kid e uma chave privada.
func NewEd25519Signer(kid string, priv ed25519.PrivateKey) *Ed25519Signer {
	return &Ed25519Signer{kid: kid, key: priv}
}

// Kid devolve o identificador da chave.
func (s *Ed25519Signer) Kid() string { return s.kid }

// Sign assina message com a chave privada.
func (s *Ed25519Signer) Sign(message []byte) ([]byte, error) {
	if len(s.key) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("chave privada inválida: %d bytes", len(s.key))
	}
	return ed25519.Sign(s.key, message), nil
}

// StaticKeys constrói um KeyLookup a partir de um mapa kid -> chave pública.
func StaticKeys(keys map[string]ed25519.PublicKey) KeyLookup {
	return func(kid string) (ed25519.PublicKey, bool) {
		k, ok := keys[kid]
		return k, ok
	}
}

// Encode serializa e comprime um replay. Com signer != nil, grava o kid no
// cabeçalho e anexa a assinatura no fim; com signer == nil, produz um arquivo
// sem assinatura (kid vazio).
func Encode(r *Replay, signer Signer) ([]byte, error) {
	kid := ""
	if signer != nil {
		kid = signer.Kid()
		if len(kid) > MaxKidBytes {
			return nil, fmt.Errorf("kid com %d bytes (máx %d)", len(kid), MaxKidBytes)
		}
	}

	body, err := encodeBody(r, kid)
	if err != nil {
		return nil, err
	}

	gz := new(bytes.Buffer)
	zw, err := gzip.NewWriterLevel(gz, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	// Header determinístico (independe de relógio/OS da máquina que grava).
	zw.Header.ModTime = time.Unix(0, 0)
	zw.Header.OS = 255
	if _, err := zw.Write(body); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	gzipped := gz.Bytes()

	out := make([]byte, 0, 4+len(gzipped)+SignatureSize)
	out = binary.BigEndian.AppendUint32(out, uint32(len(gzipped)))
	out = append(out, gzipped...)

	if signer != nil {
		sig, err := signer.Sign(signedMessage(gzipped))
		if err != nil {
			return nil, err
		}
		if len(sig) != SignatureSize {
			return nil, fmt.Errorf("assinatura com %d bytes (esperado %d)", len(sig), SignatureSize)
		}
		out = append(out, sig...)
	}

	return out, nil
}

// Decode lê e valida um arquivo .hwr. Erros estruturais devolvem (nil, ErrInvalid).
// Quando o arquivo tem assinatura mas ela não confere (kid desconhecido ou
// assinatura errada), devolve o Replay com Official=false e um erro tipado, para
// o chamador exibir "replay não oficial" sem perder o conteúdo.
func Decode(data []byte, lookup KeyLookup) (*Replay, error) {
	body, kid, signature, gzipped, err := parseEnvelope(data)
	if err != nil {
		return nil, err
	}

	r, err := decodeBody(body)
	if err != nil {
		return nil, err
	}
	r.Kid = kid
	r.Signature = signature
	r.Official = false

	if kid == "" {
		if len(signature) != 0 {
			return nil, ErrInvalid
		}
		return r, nil
	}

	if len(signature) != SignatureSize {
		return nil, ErrInvalid
	}

	// Sem chaves fornecidas, não há como validar; fica "não oficial".
	if lookup == nil {
		return r, nil
	}

	pub, ok := lookup(kid)
	if !ok {
		return r, ErrUnknownKey
	}
	if !ed25519.Verify(pub, signedMessage(gzipped), signature) {
		return r, ErrInvalidSignature
	}
	r.Official = true
	return r, nil
}

// Verify verifica apenas a assinatura de um arquivo, sem revalidar as entradas.
// Devolve nil quando a assinatura confere; ErrUnknownKey/ErrInvalidSignature nos
// demais casos; ErrInvalid para arquivo sem assinatura ou estruturalmente quebrado.
func Verify(data []byte, lookup KeyLookup) error {
	if lookup == nil {
		return ErrInvalid
	}
	_, kid, signature, gzipped, err := parseEnvelope(data)
	if err != nil {
		return err
	}
	if kid == "" || len(signature) != SignatureSize {
		return ErrInvalid
	}
	pub, ok := lookup(kid)
	if !ok {
		return ErrUnknownKey
	}
	if !ed25519.Verify(pub, signedMessage(gzipped), signature) {
		return ErrInvalidSignature
	}
	return nil
}

// signedMessage devolve "HWR1-SIG" || corpo gzipado.
func signedMessage(gzipped []byte) []byte {
	msg := make([]byte, 0, len(SignatureDomain)+len(gzipped))
	msg = append(msg, SignatureDomain...)
	msg = append(msg, gzipped...)
	return msg
}

// parseEnvelope separa o arquivo em corpo descomprimido, kid, assinatura e o
// corpo gzipado bruto (usado como mensagem da assinatura).
func parseEnvelope(data []byte) (body []byte, kid string, signature []byte, gzipped []byte, err error) {
	if len(data) > MaxFileSize {
		return nil, "", nil, nil, ErrInvalid
	}
	if len(data) < 4 {
		return nil, "", nil, nil, ErrInvalid
	}

	gzipLen := int(binary.BigEndian.Uint32(data[:4]))
	if gzipLen > MaxFileSize-4 {
		return nil, "", nil, nil, ErrInvalid
	}
	if len(data) < 4+gzipLen {
		return nil, "", nil, nil, ErrInvalid
	}

	gzipped = data[4 : 4+gzipLen]
	trailing := data[4+gzipLen:]

	body, err = gunzip(gzipped, MaxDecompressedSize)
	if err != nil {
		return nil, "", nil, nil, ErrInvalid
	}

	// Lê apenas o começo do corpo para extrair o kid (magic + versões + kid).
	kid, err = readKid(body)
	if err != nil {
		return nil, "", nil, nil, ErrInvalid
	}

	if kid == "" {
		return body, "", trailing, gzipped, nil
	}
	if len(trailing) != SignatureSize {
		return nil, "", nil, nil, ErrInvalid
	}
	return body, kid, trailing, gzipped, nil
}

// gunzip descomprime limitando a max bytes (defesa contra bomba de gzip).
func gunzip(data []byte, max int) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	gz.Multistream(false)

	out, err := io.ReadAll(io.LimitReader(gz, int64(max)+1))
	if err != nil {
		return nil, err
	}
	if len(out) > max {
		return nil, ErrInvalid
	}
	return out, nil
}

// readKid lê magic + formatVersion + engineVersion + kid do início do corpo.
func readKid(body []byte) (string, error) {
	r := bytes.NewReader(body)

	magic := make([]byte, len(Magic))
	if _, err := io.ReadFull(r, magic); err != nil {
		return "", err
	}
	if string(magic) != Magic {
		return "", ErrInvalid
	}
	if _, err := readByte(r); err != nil { // format version
		return "", err
	}
	if _, err := readU16(r); err != nil { // engine version
		return "", err
	}

	kidLen, err := readByte(r)
	if err != nil {
		return "", err
	}
	if int(kidLen) > MaxKidBytes {
		return "", ErrInvalid
	}
	kid := make([]byte, kidLen)
	if _, err := io.ReadFull(r, kid); err != nil {
		return "", err
	}
	return string(kid), nil
}

// encodeBody serializa o corpo descomprimido (magic + cabeçalho + entradas + checkpoints).
func encodeBody(r *Replay, kid string) ([]byte, error) {
	if err := validate(r, kid); err != nil {
		return nil, err
	}

	var b bytes.Buffer

	b.WriteString(Magic)
	b.WriteByte(FormatVersion)
	_ = binary.Write(&b, binary.BigEndian, uint16(r.EngineVersion))

	b.WriteByte(byte(len(kid)))
	b.WriteString(kid)

	b.Write(r.MatchID[:])
	b.Write(r.Seed[:])
	_ = binary.Write(&b, binary.BigEndian, r.StartedAtMs)

	b.WriteByte(byte(len(r.Mode)))
	b.WriteString(r.Mode)

	b.WriteByte(byte(len(r.Players)))
	for _, p := range r.Players {
		b.WriteByte(p.Index)
		b.WriteByte(byte(len(p.Name)))
		b.WriteString(p.Name)
	}

	b.WriteByte(r.Flags)

	_ = binary.Write(&b, binary.BigEndian, uint16(len(r.Entries)))
	nPlayers := len(r.Players)
	bitmapLen := bitmapBytes(nPlayers)
	for _, e := range r.Entries {
		var buf [binary.MaxVarintLen64]byte
		n := binary.PutVarint(buf[:], e.DeltaMs)
		b.Write(buf[:n])
		b.WriteByte(byte(e.Type))
		b.WriteByte(e.Player)

		var flags byte
		if len(e.Connected) > 0 {
			flags |= entryFlagConnected
		}
		if len(e.Payload) > 0 {
			flags |= entryFlagPayload
		}
		b.WriteByte(flags)

		if len(e.Connected) > 0 {
			b.Write(e.Connected[:bitmapLen])
		}
		if len(e.Payload) > 0 {
			_ = binary.Write(&b, binary.BigEndian, uint16(len(e.Payload)))
			b.Write(e.Payload)
		}
	}

	_ = binary.Write(&b, binary.BigEndian, uint16(len(r.Checkpoints)))
	for _, c := range r.Checkpoints {
		b.Write(c[:])
	}

	return b.Bytes(), nil
}

// decodeBody parseia e valida o corpo descomprimido.
func decodeBody(body []byte) (*Replay, error) {
	r := bytes.NewReader(body)

	magic := make([]byte, len(Magic))
	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, ErrInvalid
	}
	if string(magic) != Magic {
		return nil, ErrInvalid
	}

	formatVer, err := readByte(r)
	if err != nil {
		return nil, ErrInvalid
	}
	if formatVer != FormatVersion {
		return nil, ErrInvalid
	}

	engineVer, err := readU16(r)
	if err != nil {
		return nil, ErrInvalid
	}
	if engineVer < 1 {
		return nil, ErrInvalid
	}

	kidLen, err := readByte(r)
	if err != nil {
		return nil, ErrInvalid
	}
	if int(kidLen) > MaxKidBytes {
		return nil, ErrInvalid
	}
	kid := make([]byte, kidLen)
	if _, err := io.ReadFull(r, kid); err != nil {
		return nil, ErrInvalid
	}

	out := &Replay{
		FormatVersion: int(formatVer),
		EngineVersion: int(engineVer),
		Kid:           string(kid),
	}

	if _, err := io.ReadFull(r, out.MatchID[:]); err != nil {
		return nil, ErrInvalid
	}
	if _, err := io.ReadFull(r, out.Seed[:]); err != nil {
		return nil, ErrInvalid
	}
	if err := readI64(r, &out.StartedAtMs); err != nil {
		return nil, ErrInvalid
	}

	mode, err := readString8(r, MaxModeBytes)
	if err != nil {
		return nil, ErrInvalid
	}
	out.Mode = mode

	playerCount, err := readByte(r)
	if err != nil {
		return nil, ErrInvalid
	}
	if int(playerCount) < MinPlayers || int(playerCount) > MaxPlayers {
		return nil, ErrInvalid
	}
	out.Players = make([]Player, 0, playerCount)
	seenIndex := make(map[uint8]bool, playerCount)
	for i := 0; i < int(playerCount); i++ {
		idx, err := readByte(r)
		if err != nil {
			return nil, ErrInvalid
		}
		name, err := readString8(r, MaxPlayerNameBytes)
		if err != nil {
			return nil, ErrInvalid
		}
		if seenIndex[idx] {
			return nil, ErrInvalid
		}
		seenIndex[idx] = true
		out.Players = append(out.Players, Player{Index: idx, Name: name})
	}

	flags, err := readByte(r)
	if err != nil {
		return nil, ErrInvalid
	}
	out.Flags = flags

	entryCount, err := readU16(r)
	if err != nil {
		return nil, ErrInvalid
	}
	if int(entryCount) < 1 || int(entryCount) > MaxEntries {
		return nil, ErrInvalid
	}
	out.Entries = make([]Entry, 0, entryCount)

	bitmapLen := bitmapBytes(len(out.Players))
	var totalMs int64
	for i := 0; i < int(entryCount); i++ {
		delta, err := binary.ReadVarint(r)
		if err != nil {
			return nil, ErrInvalid
		}
		if delta < 0 {
			return nil, ErrInvalid
		}
		totalMs += delta
		if totalMs > MaxDurationMs {
			return nil, ErrInvalid
		}

		typ, err := readByte(r)
		if err != nil {
			return nil, ErrInvalid
		}
		if typ > uint8(EntryEnd) {
			return nil, ErrInvalid
		}

		player, err := readByte(r)
		if err != nil {
			return nil, ErrInvalid
		}
		if int(player) >= len(out.Players) && player != NoPlayer {
			return nil, ErrInvalid
		}

		flags, err := readByte(r)
		if err != nil {
			return nil, ErrInvalid
		}

		e := Entry{DeltaMs: delta, Type: EntryType(typ), Player: player}

		if flags&entryFlagConnected != 0 {
			e.Connected = make([]byte, bitmapLen)
			if _, err := io.ReadFull(r, e.Connected); err != nil {
				return nil, ErrInvalid
			}
		}
		if flags&entryFlagPayload != 0 {
			payload, err := readBytes16(r, MaxPayloadBytes)
			if err != nil {
				return nil, ErrInvalid
			}
			e.Payload = payload
		}

		out.Entries = append(out.Entries, e)
	}

	checkpointCount, err := readU16(r)
	if err != nil {
		return nil, ErrInvalid
	}
	want := checkpointCountFor(len(out.Entries))
	if int(checkpointCount) != want {
		return nil, ErrInvalid
	}
	out.Checkpoints = make([]Checkpoint, 0, checkpointCount)
	for i := 0; i < int(checkpointCount); i++ {
		var c Checkpoint
		if _, err := io.ReadFull(r, c[:]); err != nil {
			return nil, ErrInvalid
		}
		out.Checkpoints = append(out.Checkpoints, c)
	}

	// Corpo não pode ter bytes sobrando.
	if r.Len() != 0 {
		return nil, ErrInvalid
	}

	return out, nil
}

// validate valida um Replay para Encode (erros descritivos; chamador confiável).
func validate(r *Replay, kid string) error {
	if r == nil {
		return errors.New("replay nil")
	}
	if r.EngineVersion < 1 {
		return fmt.Errorf("versão do motor inválida: %d", r.EngineVersion)
	}
	if len(kid) > MaxKidBytes {
		return fmt.Errorf("kid com %d bytes (máx %d)", len(kid), MaxKidBytes)
	}
	if len(r.Mode) == 0 || len(r.Mode) > MaxModeBytes {
		return fmt.Errorf("modo com %d bytes (1..%d)", len(r.Mode), MaxModeBytes)
	}
	if len(r.Players) < MinPlayers || len(r.Players) > MaxPlayers {
		return fmt.Errorf("%d jogadores (esperado %d..%d)", len(r.Players), MinPlayers, MaxPlayers)
	}
	seen := make(map[uint8]bool, len(r.Players))
	for _, p := range r.Players {
		if len(p.Name) == 0 || len(p.Name) > MaxPlayerNameBytes {
			return fmt.Errorf("nome de jogador com %d bytes (1..%d)", len(p.Name), MaxPlayerNameBytes)
		}
		if seen[p.Index] {
			return fmt.Errorf("índice de jogador duplicado: %d", p.Index)
		}
		seen[p.Index] = true
	}
	if len(r.Entries) < 1 || len(r.Entries) > MaxEntries {
		return fmt.Errorf("%d entradas (esperado 1..%d)", len(r.Entries), MaxEntries)
	}

	bitmapLen := bitmapBytes(len(r.Players))
	var totalMs int64
	for i, e := range r.Entries {
		if e.DeltaMs < 0 {
			return fmt.Errorf("entrada %d com delta negativo", i)
		}
		totalMs += e.DeltaMs
		if totalMs > MaxDurationMs {
			return fmt.Errorf("duração excede %d ms", MaxDurationMs)
		}
		if e.Type > EntryEnd {
			return fmt.Errorf("entrada %d com tipo inválido: %d", i, e.Type)
		}
		if int(e.Player) >= len(r.Players) && e.Player != NoPlayer {
			return fmt.Errorf("entrada %d com índice de jogador inválido: %d", i, e.Player)
		}
		if e.Connected != nil && len(e.Connected) != bitmapLen {
			return fmt.Errorf("entrada %d com bitmap de %d bytes (esperado %d)", i, len(e.Connected), bitmapLen)
		}
		if len(e.Payload) > MaxPayloadBytes {
			return fmt.Errorf("entrada %d com payload de %d bytes (máx %d)", i, len(e.Payload), MaxPayloadBytes)
		}
	}

	want := checkpointCountFor(len(r.Entries))
	if len(r.Checkpoints) != want {
		return fmt.Errorf("%d checkpoints (esperado %d)", len(r.Checkpoints), want)
	}

	return nil
}

// checkpointCountFor devolve ceil(n/CheckpointInterval): um a cada bloco de 20
// e um no fim.
func checkpointCountFor(n int) int {
	return (n + CheckpointInterval - 1) / CheckpointInterval
}

func bitmapBytes(nPlayers int) int {
	return (nPlayers + 7) / 8
}

// --- leitores auxiliares (big-endian) ---

func readByte(r *bytes.Reader) (byte, error) {
	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	return b, nil
}

func readU16(r *bytes.Reader) (uint16, error) {
	var b [2]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b[:]), nil
}

func readI64(r *bytes.Reader, out *int64) error {
	var b [8]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return err
	}
	*out = int64(binary.BigEndian.Uint64(b[:]))
	return nil
}

func readString8(r *bytes.Reader, max int) (string, error) {
	n, err := readByte(r)
	if err != nil {
		return "", err
	}
	if int(n) > max {
		return "", ErrInvalid
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	return string(b), nil
}

func readBytes16(r *bytes.Reader, max int) ([]byte, error) {
	n, err := readU16(r)
	if err != nil {
		return nil, err
	}
	if int(n) > max {
		return nil, ErrInvalid
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}
	return b, nil
}
