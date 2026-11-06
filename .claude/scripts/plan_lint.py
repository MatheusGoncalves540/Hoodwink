#!/usr/bin/env python3
"""Lint estrutural de planos (./plans). Custo zero de tokens.
Uso: python3 .claude/scripts/plan_lint.py plans/01-x.md [mais planos...]
Saída: uma linha por problema + resumo. Código de saída 1 se houver ERRO.
"""
import json
import re
import sys
from pathlib import Path

PAPEIS = {"Arquiteto", "Backend", "Frontend", "QA", "DevOps", "Seg", "Docs"}
CAMPOS = ["Status:", "Objetivo:", "Fora do escopo:", "Premissas e decisões:", "Pré-requisitos:"]
SECOES = ["## Tarefas", "## Aceite final", "## Riscos"]
TAREFA = re.compile(r"^- \[([ x!])\] (T\d+) \[(\w+)\] (.+)$")
VAGO = re.compile(r"^(funciona|ok|testar|validar|revisar|conferir|checar|deve funcionar)\b", re.I)
CRASE = re.compile(r"`([^`]+)`")
NPM = re.compile(r"\b(?:npm|pnpm|yarn|bun)\s+(?:run\s+)?([\w:.-]+)")
PATH = re.compile(r"(?:^|\s)((?:\./)?[\w.-]+(?:/[\w.\-\[\]]+)+\.\w+)")


def raiz(p: Path) -> Path:
    for d in [p.resolve().parent, *p.resolve().parents]:
        if any((d / m).exists() for m in (".git", "package.json", "pyproject.toml")):
            return d
    return Path.cwd()


def scripts_npm(r: Path):
    try:
        return set(json.loads((r / "package.json").read_text()).get("scripts", {}))
    except Exception:
        return None


def secao(linhas, titulo):
    out, dentro = [], False
    for l in linhas:
        if l.startswith("## "):
            dentro = l.strip().startswith(titulo)
            continue
        if dentro:
            out.append(l)
    return out


def caminho_critico(deps):
    memo = {}

    def prof(t, pilha=()):
        if t in pilha:
            return 0
        if t not in memo:
            memo[t] = 1 + max((prof(d, pilha + (t,)) for d in deps.get(t, []) if d in deps), default=0)
        return memo[t]

    return max((prof(t) for t in deps), default=0)


def ciclos(deps):
    cor, achados = {}, []

    def dfs(t, pilha):
        cor[t] = 1
        for d in deps.get(t, []):
            if d not in deps:
                continue
            if cor.get(d) == 1:
                achados.append(" → ".join(pilha[pilha.index(d):] + [d]))
            elif not cor.get(d):
                dfs(d, pilha + [d])
        cor[t] = 2

    for t in deps:
        if not cor.get(t):
            dfs(t, [t])
    return achados


def alcanca(deps, a, b):
    vistos, fila = set(), [a]
    while fila:
        x = fila.pop()
        for d in deps.get(x, []):
            if d == b:
                return True
            if d not in vistos:
                vistos.add(d)
                fila.append(d)
    return False


def lint(arquivo: Path):
    probs = []
    E = lambda m: probs.append(("ERRO", m))
    A = lambda m: probs.append(("AVISO", m))
    txt = arquivo.read_text(encoding="utf-8")
    linhas = txt.splitlines()
    r = raiz(arquivo)
    npm = scripts_npm(r)

    if not linhas or not linhas[0].startswith("# PLANO"):
        E("cabeçalho '# PLANO NN: título' ausente na 1ª linha")
    for c in CAMPOS:
        l = next((l for l in linhas if l.startswith(c)), None)
        if l is None:
            E(f"campo '{c}' ausente")
        elif not l[len(c):].strip():
            E(f"campo '{c}' vazio")
    for s in SECOES:
        if not any(l.strip().startswith(s) for l in linhas):
            E(f"seção '{s}' ausente")
    if not any(l.startswith("Esforço:") for l in linhas):
        A("'Esforço:' ausente")

    tarefas, deps, arqs_por_tarefa, papeis = {}, {}, {}, set()
    bloco = secao(linhas, "## Tarefas")
    for l in bloco:
        if not l.startswith("- ["):
            continue
        m = TAREFA.match(l)
        if not m:
            E(f"tarefa fora do formato: {l[:70]}")
            continue
        st, tid, papel, resto = m.groups()
        if tid in tarefas:
            E(f"{tid}: ID duplicado")
        tarefas[tid] = st
        papeis.add(papel)
        if papel not in PAPEIS:
            E(f"{tid}: papel '{papel}' fora do catálogo {sorted(PAPEIS)}")
        partes = [p.strip() for p in resto.split(" · ")]
        desc = partes[0]
        campos = {p.split(":", 1)[0].strip(): p.split(":", 1)[1].strip() for p in partes[1:] if ":" in p}
        if len(desc) > 160:
            A(f"{tid}: descrição longa ({len(desc)} chars), provavelmente são 2 tarefas")
        arqs = CRASE.findall(campos.get("arquivos", ""))
        arqs_por_tarefa[tid] = set(arqs)
        if not arqs:
            A(f"{tid}: sem 'arquivos'; o executor vai precisar procurar (gasta tokens)")
        for a in arqs:
            p = r / a
            if not p.exists() and not p.parent.exists():
                A(f"{tid}: `{a}` não existe e a pasta também não; confira o caminho")
        d = campos.get("depende", "—")
        deps[tid] = re.findall(r"T\d+", d)
        v = campos.get("verificar", "")
        if not v:
            E(f"{tid}: sem 'verificar'")
        elif VAGO.match(v.strip("` ")):
            A(f"{tid}: 'verificar' vago ({v[:40]}); use um comando ou um resultado observável")
        for cmd in CRASE.findall(v):
            for s in NPM.findall(cmd):
                if npm is not None and s not in npm and s not in {"install", "i", "test", "exec", "dlx", "x"}:
                    E(f"{tid}: script '{s}' não existe no package.json")
            for pth in PATH.findall(cmd):
                if not (r / pth).exists() and not any(pth in arqs_por_tarefa[t] for t in arqs_por_tarefa):
                    A(f"{tid}: 'verificar' usa `{pth}`, que não existe e nenhuma tarefa cria")

    if not tarefas:
        E("nenhuma tarefa encontrada")
    ids = list(tarefas)
    for t, ds in deps.items():
        for d in ds:
            if d == t:
                E(f"{t}: depende de si mesma")
            elif d not in tarefas:
                E(f"{t}: depende de {d}, que não existe")
            elif ids.index(d) > ids.index(t):
                A(f"{t}: depende de {d}, que vem depois; reordene")
    for c in ciclos(deps):
        E(f"ciclo de dependência: {c}")
    for i, a in enumerate(ids):
        for b in ids[i + 1:]:
            comum = arqs_por_tarefa[a] & arqs_por_tarefa[b]
            if comum and not alcanca(deps, a, b) and not alcanca(deps, b, a):
                A(f"{a} e {b} mexem em {sorted(comum)[0]} sem dependência entre si (conflito se rodarem em paralelo)")

    n = len(tarefas)
    if n > 8 and not any(l.startswith("### Fase") for l in bloco):
        A(f"{n} tarefas sem agrupamento em fases")
    if n > 25:
        A(f"{n} tarefas; considere dividir em mais de um plano")
    if "Frontend" in papeis:
        pre = next((l for l in linhas if l.startswith("Pré-requisitos:")), "")
        if "UI" not in pre and "ui" not in pre.lower().split():
            A("há tarefa de Frontend, mas Pré-requisitos não aponta a referência de UI")
    if n > 5 and "Docs" not in papeis:
        A("plano com mais de 5 tarefas e nenhuma tarefa [Docs]")

    aceite = [l for l in secao(linhas, "## Aceite final") if l.startswith("- [")]
    if not aceite:
        E("'Aceite final' sem critérios")
    sem_cmd = [l for l in aceite if "`" not in l]
    if sem_cmd:
        A(f"{len(sem_cmd)} critério(s) de aceite sem comando executável")
    riscos = [l for l in secao(linhas, "## Riscos") if l.startswith("- ")]
    if len(riscos) > 5:
        A(f"{len(riscos)} riscos; deixe só os 5 reais")

    feitas = sum(1 for s in tarefas.values() if s == "x")
    resumo = f"{n} tarefas ({feitas} feitas) · caminho crítico {caminho_critico(deps)}/{n}"
    return probs, resumo


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        sys.exit(2)
    falhou = False
    for a in sys.argv[1:]:
        probs, resumo = lint(Path(a))
        for nivel, m in probs:
            print(f"{nivel} {m}")
        ne = sum(1 for n, _ in probs if n == "ERRO")
        na = len(probs) - ne
        print(f"{a}: {ne} erros, {na} avisos · {resumo}")
        falhou |= ne > 0
    sys.exit(1 if falhou else 0)


if __name__ == "__main__":
    main()
