/**
 * Aba "Bugs" do super-admin (plano 06, T7), aberta por `#admin`.
 * Relatos vêm da API e são exibidos como texto (`<pre>{JSON.stringify(...)}</pre>`): nunca HTML.
 * 403 mostra "Sem acesso"; 401 pede login de novo.
 */
import { useEffect, useId, useState } from "react";
import { useLayout } from "../../app/useLayout";
import { BUG_CATEGORY_LABELS } from "../../game/copy";
import {
  AdminApiError,
  BUG_SOURCES,
  BUG_STATUSES,
  listBugReports,
  updateBugReportStatus,
  type AdminBugReport,
  type AdminBugReportsResponse,
  type BugReportSource,
  type BugReportStatus,
} from "../../net/adminApi";
import { useAuthStore } from "../../store/authStore";
import { Button } from "../../ui/Button/Button";
import { Sheet } from "../../ui/Sheet/Sheet";
import { TextField } from "../../ui/TextField/TextField";
import { cx } from "../../ui/cx";
import { isUuid, pageCount } from "./filters";
import s from "./AdminBugs.module.css";

const LIMIT = 50;

const t = {
  lead: "Relatos enviados pelo Lobby e marcados durante as partidas.",
  filters: "Filtros de relatos",
  status: "Status",
  allStatuses: "Todos",
  source: "Origem",
  allSources: "Todas",
  matchId: "Partida",
  matchPlaceholder: "UUID da partida",
  applyMatch: "Filtrar",
  invalidMatch: "Informe um UUID válido.",
  reset: "Limpar filtros",
  retry: "Tentar de novo",
  loading: "Carregando relatos…",
  emptyTitle: "Nenhum relato encontrado.",
  emptyHint: "Ajuste os filtros ou volte mais tarde.",
  errorTitle: "Não foi possível carregar os relatos.",
  unexpectedError: "Erro inesperado. Tente de novo.",
  forbiddenTitle: "Sem acesso",
  forbiddenHint: "Sua conta não é super-admin.",
  signedOutTitle: "Entre para ver os relatos.",
  signedOutHint: "Faça login no lobby e abra #admin de novo.",
  expiredTitle: "Sua sessão expirou.",
  expiredHint: "Entre de novo pelo lobby.",
  back: "☜ Voltar ao início",
  previous: "Anterior",
  next: "Próxima",
  pageLabel: (page: number, pages: number) => `Página ${page} de ${pages}`,
  totalLabel: (total: number) => (total === 1 ? "1 relato" : `${total} relatos`),
  novosLabel: (novos: number) => (novos === 1 ? "1 novo" : `${novos} novos`),
  tableCaption: "Relatos de bug, do mais novo para o mais antigo",
  results: "Resultados",
  received: "Recebido",
  statusField: "Status",
  category: "Categoria",
  usernameField: "Usuário",
  turns: "Turnos",
  matchIdField: "Partida",
  details: "Detalhes",
  detailsOf: (id: number) => `Detalhes do relato ${id}`,
  pagination: "Paginação",
  close: "☜ Fechar",
  detailTitle: (id: number) => `Relato ${id}`,
  updatedAt: "Atualizado",
  sourceField: "Origem",
  userIdField: "ID do usuário",
  anonymous: "Anônimo",
  none: "-",
  noCategory: "Sem categoria",
  description: "Descrição",
  noDescription: "Sem descrição.",
  context: "Contexto",
  noContext: "Sem contexto.",
  userAgent: "User-Agent",
  noUserAgent: "Sem User-Agent.",
  ip: "IP",
  noIp: "Sem IP.",
  changeStatus: "Trocar status",
  saving: "Salvando…",
  saved: "Status atualizado.",
  gone: "Esse relato não existe mais.",
  statuses: { novo: "Novo", visto: "Visto", resolvido: "Resolvido" } satisfies Record<BugReportStatus, string>,
  actions: {
    novo: "Marcar como novo",
    visto: "Marcar como visto",
    resolvido: "Marcar como resolvido",
  } satisfies Record<BugReportStatus, string>,
  sources: { partida: "Partida", lobby: "Lobby" } satisfies Record<BugReportSource, string>,
};

type View =
  | { kind: "loading" }
  | { kind: "ready"; data: AdminBugReportsResponse }
  | { kind: "error"; message: string }
  | { kind: "forbidden" }
  | { kind: "unauthorized" };

const dateTime = new Intl.DateTimeFormat("pt-BR", { dateStyle: "short", timeStyle: "medium" });

const formatAt = (iso: string) => {
  const date = new Date(iso);
  return Number.isNaN(date.getTime()) ? iso : dateTime.format(date);
};

const statusClass = (status: BugReportStatus) =>
  status === "novo" ? s.statusNovo : status === "resolvido" ? s.statusResolvido : s.statusVisto;

const turnsLabel = (turns: number[]) => (turns.length ? turns.join(", ") : t.none);

/** Detalhe: descrição, contexto, User-Agent e IP sempre em `<pre>` (texto puro, sem HTML). */
export function Detail({
  item,
  saving,
  saveError,
  notice,
  onStatus,
}: {
  item: AdminBugReport;
  saving: BugReportStatus | null;
  saveError: string | null;
  notice: string | null;
  onStatus(status: BugReportStatus): void;
}) {
  const rows: [string, string][] = [
    [t.received, formatAt(item.created_at)],
    [t.updatedAt, formatAt(item.updated_at)],
    [t.sourceField, t.sources[item.source]],
    [t.statusField, t.statuses[item.status]],
    [t.usernameField, item.username ?? t.anonymous],
    [t.userIdField, item.user_id ?? t.none],
    [t.matchIdField, item.match_id ?? t.none],
    [t.turns, turnsLabel(item.turns)],
    [t.category, item.category ? BUG_CATEGORY_LABELS[item.category] : t.noCategory],
  ];
  const context = item.context == null ? t.noContext : JSON.stringify(item.context, null, 2);

  return (
    <article className={s.detail}>
      <h2 className={s.detailTitle}>{t.detailTitle(item.id)}</h2>
      <dl className={s.meta}>
        {rows.map(([label, value]) => (
          <div key={label} className={s.metaRow}>
            <dt className={cx("hw-rotulo", s.metaLabel)}>{label}</dt>
            <dd className={s.metaValue}>{value}</dd>
          </div>
        ))}
      </dl>
      <h3 className={s.detailHeading}>{t.description}</h3>
      <pre className={s.pre}>{item.description ?? t.noDescription}</pre>
      <h3 className={s.detailHeading}>{t.context}</h3>
      <pre className={s.pre}>{context}</pre>
      <h3 className={s.detailHeading}>{t.userAgent}</h3>
      <pre className={s.pre}>{item.user_agent ?? t.noUserAgent}</pre>
      <h3 className={s.detailHeading}>{t.ip}</h3>
      <pre className={s.pre}>{item.ip ?? t.noIp}</pre>
      <h3 className={s.detailHeading}>{t.changeStatus}</h3>
      <div className={s.statusActions}>
        {BUG_STATUSES.map((status) => (
          <Button
            key={status}
            variant="outline"
            disabled={!!saving || item.status === status}
            aria-pressed={item.status === status}
            onClick={() => onStatus(status)}
          >
            {saving === status ? t.saving : t.actions[status]}
          </Button>
        ))}
      </div>
      {saving && (
        <p className={s.note} role="status">
          {t.saving}
        </p>
      )}
      {saveError && (
        <p className={s.errorText} role="alert">
          {saveError}
        </p>
      )}
      {notice && !saving && (
        <p className={s.note} role="status">
          {notice}
        </p>
      )}
    </article>
  );
}

type Props = {
  /** A aba está visível? Fechamos a gaveta ao trocar de aba. */
  active: boolean;
  /** Sobe o contador de `novos` para a aba. */
  onNovosChange(novos: number): void;
};

export function AdminBugs({ active, onNovosChange }: Props) {
  const layout = useLayout();
  const token = useAuthStore((st) => st.jwtToken);
  const ids = useId();
  const matchErrorId = `${ids}-match`;

  const [status, setStatus] = useState<BugReportStatus | "">("");
  const [source, setSource] = useState<BugReportSource | "">("");
  const [matchDraft, setMatchDraft] = useState("");
  const [matchId, setMatchId] = useState("");
  const [matchError, setMatchError] = useState("");
  const [page, setPage] = useState(1);
  const [reloadKey, setReloadKey] = useState(0);
  const [view, setView] = useState<View>({ kind: "loading" });
  const [selected, setSelected] = useState<AdminBugReport | null>(null);
  const [saving, setSaving] = useState<{ id: number; status: BugReportStatus } | null>(null);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  useEffect(() => {
    if (active) return;
    setSelected(null);
    setSaveError(null);
    setNotice(null);
  }, [active]);

  useEffect(() => {
    if (!token) {
      setView({ kind: "unauthorized" });
      setSelected(null);
      return;
    }

    const controller = new AbortController();
    setView({ kind: "loading" });

    listBugReports(
      token,
      { status, source, matchId: matchId || undefined, page, limit: LIMIT },
      controller.signal,
    )
      .then((data) => {
        setView({ kind: "ready", data });
        onNovosChange(data.novos);
      })
      .catch((error: unknown) => {
        if (controller.signal.aborted) return;
        if (error instanceof AdminApiError) {
          if (error.status === 403) {
            setSelected(null);
            return setView({ kind: "forbidden" });
          }
          if (error.status === 401) {
            setSelected(null);
            return setView({ kind: "unauthorized" });
          }
          return setView({ kind: "error", message: error.message });
        }
        if ((error as { name?: string } | null)?.name === "AbortError") return;
        setView({ kind: "error", message: t.unexpectedError });
      });

    return () => controller.abort();
  }, [token, status, source, matchId, page, reloadKey, onNovosChange]);

  const changeStatus = (value: BugReportStatus | "") => {
    setStatus(value);
    setPage(1);
    setSelected(null);
    setSaveError(null);
    setNotice(null);
  };

  const changeSource = (value: BugReportSource | "") => {
    setSource(value);
    setPage(1);
    setSelected(null);
    setSaveError(null);
    setNotice(null);
  };

  const applyMatch = () => {
    const value = matchDraft.trim();
    if (value && !isUuid(value)) {
      setMatchError(t.invalidMatch);
      return;
    }
    setMatchError("");
    setMatchId(value.toLowerCase());
    setPage(1);
    setSelected(null);
    setSaveError(null);
    setNotice(null);
  };

  const reset = () => {
    setStatus("");
    setSource("");
    setMatchDraft("");
    setMatchId("");
    setMatchError("");
    setPage(1);
    setSelected(null);
    setSaveError(null);
    setNotice(null);
  };

  const saveStatus = async (next: BugReportStatus) => {
    if (!token || !selected || saving) return;
    setSaving({ id: selected.id, status: next });
    setSaveError(null);
    setNotice(null);

    try {
      const updated = await updateBugReportStatus(token, selected.id, next);
      setSelected(updated);
      setNotice(t.saved);
      setReloadKey((n) => n + 1);
    } catch (error) {
      if (error instanceof AdminApiError) {
        if (error.status === 401) {
          setView({ kind: "unauthorized" });
          setSelected(null);
          return;
        }
        if (error.status === 403) {
          setView({ kind: "forbidden" });
          setSelected(null);
          return;
        }
        setSaveError(error.status === 404 ? t.gone : error.message);
      } else {
        setSaveError(t.unexpectedError);
      }
    } finally {
      setSaving(null);
    }
  };

  const totalPages = view.kind === "ready" ? pageCount(view.data.total, view.data.limit) : 1;
  const busy = view.kind === "loading";
  const savingStatus = selected && saving?.id === selected.id ? saving.status : null;

  // Trocar o status pode esvaziar a página atual; volta para a última que existe.
  useEffect(() => {
    if (view.kind !== "ready") return;
    const pages = pageCount(view.data.total, view.data.limit);
    if (page > pages) setPage(pages);
  }, [view, page]);

  return (
    <div className={s.bugs}>
      <p className={s.lead}>{t.lead}</p>

      <section aria-label={t.filters} className={s.filters}>
        <label className={s.field}>
          <span className={cx("hw-rotulo", s.caption)}>{t.status}</span>
          <select
            className={s.control}
            value={status}
            onChange={(e) => changeStatus(e.target.value as BugReportStatus | "")}
          >
            <option value="">{t.allStatuses}</option>
            {BUG_STATUSES.map((value) => (
              <option key={value} value={value}>
                {t.statuses[value]}
              </option>
            ))}
          </select>
        </label>

        <label className={s.field}>
          <span className={cx("hw-rotulo", s.caption)}>{t.source}</span>
          <select
            className={s.control}
            value={source}
            onChange={(e) => changeSource(e.target.value as BugReportSource | "")}
          >
            <option value="">{t.allSources}</option>
            {BUG_SOURCES.map((value) => (
              <option key={value} value={value}>
                {t.sources[value]}
              </option>
            ))}
          </select>
        </label>

        <div className={s.fieldWide}>
          <span aria-hidden="true" className={cx("hw-rotulo", s.caption)}>
            {t.matchId}
          </span>
          <TextField
            label={t.matchId}
            placeholder={t.matchPlaceholder}
            value={matchDraft}
            onChange={setMatchDraft}
            onSubmit={applyMatch}
            action={t.applyMatch}
            maxLength={36}
            autoCapitalize="none"
            aria-invalid={matchError ? true : undefined}
            aria-describedby={matchError ? matchErrorId : undefined}
          />
        </div>

        <div className={s.actions}>
          <Button variant="link" onClick={reset}>
            {t.reset}
          </Button>
        </div>

        {matchError && (
          <p id={matchErrorId} role="alert" className={s.errorText}>
            {matchError}
          </p>
        )}
      </section>

      <section aria-label={t.results} aria-busy={busy} className={s.results}>
        {view.kind === "loading" && (
          <p className={s.state} role="status">
            {t.loading}
          </p>
        )}

        {view.kind === "unauthorized" && (
          <div className={s.state}>
            <h2 className={s.stateTitle}>{token ? t.expiredTitle : t.signedOutTitle}</h2>
            <p>{token ? t.expiredHint : t.signedOutHint}</p>
            <a className={s.stateLink} href="#">
              {t.back}
            </a>
          </div>
        )}

        {view.kind === "forbidden" && (
          <div className={s.state} role="alert">
            <h2 className={s.stateTitle}>{t.forbiddenTitle}</h2>
            <p>{t.forbiddenHint}</p>
          </div>
        )}

        {view.kind === "error" && (
          <div className={s.state} role="alert">
            <h2 className={s.stateTitle}>{t.errorTitle}</h2>
            <p>{view.message}</p>
            <Button variant="outline" onClick={() => setReloadKey((n) => n + 1)}>
              {t.retry}
            </Button>
          </div>
        )}

        {view.kind === "ready" && view.data.items.length === 0 && (
          <div className={s.state}>
            <h2 className={s.stateTitle}>{t.emptyTitle}</h2>
            <p>{t.emptyHint}</p>
          </div>
        )}

        {view.kind === "ready" && view.data.items.length > 0 && (
          <>
            <div className={s.summary} role="status">
              <span>{t.totalLabel(view.data.total)}</span>
              <span>{t.novosLabel(view.data.novos)}</span>
              <span>{t.pageLabel(page, totalPages)}</span>
            </div>

            <div className={s.tableWrap}>
              <table className={s.table}>
                <caption className={s.srOnly}>{t.tableCaption}</caption>
                <thead>
                  <tr>
                    <th scope="col">{t.received}</th>
                    <th scope="col">{t.sourceField}</th>
                    <th scope="col">{t.statusField}</th>
                    <th scope="col">{t.category}</th>
                    <th scope="col">{t.usernameField}</th>
                    <th scope="col">{t.turns}</th>
                    <th scope="col">{t.matchIdField}</th>
                    <th scope="col">
                      <span className={s.srOnly}>{t.details}</span>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {view.data.items.map((item) => (
                    <tr key={item.id} className={cx(item.status === "novo" && s.rowNew)}>
                      <td className={s.nowrap}>{formatAt(item.created_at)}</td>
                      <td>{t.sources[item.source]}</td>
                      <td>
                        <span className={cx(s.status, statusClass(item.status))}>{t.statuses[item.status]}</span>
                      </td>
                      <td>{item.category ? BUG_CATEGORY_LABELS[item.category] : t.none}</td>
                      <td>{item.username ?? item.user_id ?? t.anonymous}</td>
                      <td className={s.nowrap}>{turnsLabel(item.turns)}</td>
                      <td className={s.mono}>{item.match_id ?? t.none}</td>
                      <td>
                        <Button
                          variant="link"
                          className={s.details}
                          aria-label={t.detailsOf(item.id)}
                          onClick={() => {
                            setSaveError(null);
                            setNotice(null);
                            setSelected(item);
                          }}
                        >
                          {t.details}
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <nav aria-label={t.pagination} className={s.pager}>
              <Button variant="outline" disabled={page <= 1 || busy} onClick={() => setPage((n) => Math.max(1, n - 1))}>
                {t.previous}
              </Button>
              <span className={s.pageNow}>{t.pageLabel(page, totalPages)}</span>
              <Button
                variant="outline"
                disabled={page >= totalPages || busy}
                onClick={() => setPage((n) => Math.min(totalPages, n + 1))}
              >
                {t.next}
              </Button>
            </nav>
          </>
        )}
      </section>

      <Sheet
        open={!!selected && active}
        onClose={() => {
          setSelected(null);
          setSaveError(null);
          setNotice(null);
        }}
        label={selected ? t.detailTitle(selected.id) : t.details}
        mode={layout}
        backLabel={t.close}
        fill
      >
        {selected && (
          <Detail
            item={selected}
            saving={savingStatus}
            saveError={saveError}
            notice={notice}
            onStatus={saveStatus}
          />
        )}
      </Sheet>
    </div>
  );
}
