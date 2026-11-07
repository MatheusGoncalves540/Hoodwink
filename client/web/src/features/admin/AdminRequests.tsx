/**
 * Tela de requests do super-admin (plano 04, T8), aberta por `#admin` sem link no menu.
 * Tudo que vem da API é renderizado como texto (`{valor}` ou `<pre>`): nunca HTML.
 * 403 mostra "Sem acesso"; 401 pede login de novo.
 */
import { useEffect, useId, useRef, useState, type KeyboardEvent } from "react";
import { useLayout } from "../../app/useLayout";
import {
  AdminApiError,
  listRequestLogs,
  type AdminRequestLog,
  type AdminRequestLogsResponse,
} from "../../net/adminApi";
import { useAuthStore } from "../../store/authStore";
import { Button } from "../../ui/Button/Button";
import { Sheet } from "../../ui/Sheet/Sheet";
import { TextField } from "../../ui/TextField/TextField";
import { Toggle } from "../../ui/Toggle/Toggle";
import { cx } from "../../ui/cx";
import { AdminBugs } from "./AdminBugs";
import {
  PAGE_SIZE_OPTIONS,
  PERIODS,
  clampLimit,
  customRange,
  isUuid,
  pageCount,
  parseStatusInput,
  presetRange,
  toLocalInput,
  type Period,
} from "./filters";
import s from "./AdminRequests.module.css";

const t = {
  title: "Requisições",
  lead: "Chamadas HTTP recebidas pelo backend e pelo gameServer.",
  tabs: "Seções do admin",
  tabRequests: "Requisições",
  tabBugs: "Bugs",
  bugsWithNew: (novos: number) => (novos === 1 ? "Bugs, 1 novo" : `Bugs, ${novos} novos`),
  bugsTitle: "Bugs",
  bugsLead: "Relatos de bug enviados pelo Lobby e marcados durante as partidas.",
  filters: "Filtros",
  results: "Resultados",
  route: "Rota",
  routePlaceholder: "/getTicket",
  period: "Período",
  from: "De",
  to: "Até",
  status: "Status",
  statusPlaceholder: "4xx ou 200",
  username: "Nome do usuário",
  usernamePlaceholder: "Nome do usuário",
  userId: "ID do usuário",
  userIdPlaceholder: "UUID do usuário",
  service: "Serviço",
  allServices: "Todos",
  limit: "Por página",
  archived: "Arquivados",
  archivedHint: "Requisições com mais de 30 dias. O período pode ir até 183 dias.",
  apply: "Aplicar filtros",
  reset: "Limpar filtros",
  retry: "Tentar de novo",
  loading: "Carregando requisições…",
  emptyTitle: "Nenhuma requisição encontrada.",
  emptyHint: "Ajuste os filtros ou aumente o período.",
  errorTitle: "Não foi possível carregar as requisições.",
  unexpectedError: "Erro inesperado. Tente de novo.",
  forbiddenTitle: "Sem acesso",
  forbiddenHint: "Sua conta não é super-admin.",
  signedOutTitle: "Entre para ver as requisições.",
  signedOutHint: "Faça login no lobby e abra #admin de novo.",
  expiredTitle: "Sua sessão expirou.",
  expiredHint: "Entre de novo pelo lobby.",
  back: "☜ Voltar ao início",
  previous: "Anterior",
  next: "Próxima",
  pageLabel: (page: number, pages: number) => `Página ${page} de ${pages}`,
  totalLabel: (total: number) => (total === 1 ? "1 requisição" : `${total} requisições`),
  tableCaption: "Requisições recebidas",
  details: "Detalhes",
  detailsOf: (id: number) => `Detalhes da requisição ${id}`,
  pagination: "Paginação",
  detailTitle: (id: number) => `Requisição ${id}`,
  close: "☜ Fechar",
  body: "Corpo",
  errorField: "Erro",
  userAgent: "User-Agent",
  ip: "IP",
  noBody: "Sem corpo.",
  noError: "Sem erro.",
  noUserAgent: "Sem User-Agent.",
  noIp: "Sem IP.",
  at: "Horário",
  serviceField: "Serviço",
  method: "Método",
  routeField: "Rota",
  path: "Caminho",
  statusField: "Status",
  duration: "Duração",
  usernameField: "Usuário",
  userIdField: "ID do usuário",
  requestId: "Request ID",
};

type Service = "" | "backend" | "gameServer";

type Applied = {
  from: string;
  to: string;
  route: string;
  status: string;
  username: string;
  userId: string;
  service: Service;
  archived: boolean;
  limit: number;
};

type View =
  | { kind: "loading" }
  | { kind: "ready"; data: AdminRequestLogsResponse }
  | { kind: "error"; message: string }
  | { kind: "forbidden" }
  | { kind: "unauthorized" };

type Tab = "requests" | "bugs";

const dateTime = new Intl.DateTimeFormat("pt-BR", { dateStyle: "short", timeStyle: "medium" });

const formatAt = (iso: string) => {
  const date = new Date(iso);
  return Number.isNaN(date.getTime()) ? iso : dateTime.format(date);
};

const statusClass = (status: number | null) => {
  if (status == null) return s.statusNone;
  if (status >= 500) return s.statusError;
  if (status >= 400) return s.statusWarn;
  return s.statusOk;
};

const emptyApplied = (): Applied => {
  const range = presetRange("24h");
  return {
    ...range,
    route: "",
    status: "",
    username: "",
    userId: "",
    service: "",
    archived: false,
    limit: PAGE_SIZE_OPTIONS[0],
  };
};

/** Detalhe: corpo, erro, User-Agent e IP sempre em `<pre>` (texto puro, sem HTML). */
function Detail({ item }: { item: AdminRequestLog }) {
  const rows: [string, string][] = [
    [t.at, formatAt(item.at)],
    [t.serviceField, item.service],
    [t.method, item.method ?? "-"],
    [t.routeField, item.route ?? "-"],
    [t.path, item.path ?? "-"],
    [t.statusField, item.status == null ? "-" : String(item.status)],
    [t.duration, item.duration_ms == null ? "-" : `${item.duration_ms} ms`],
    [t.usernameField, item.username ?? "-"],
    [t.userIdField, item.user_id ?? "-"],
    [t.requestId, item.request_id ?? "-"],
  ];

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
      <h3 className={s.detailHeading}>{t.body}</h3>
      <pre className={s.pre}>{item.body ?? t.noBody}</pre>
      <h3 className={s.detailHeading}>{t.errorField}</h3>
      <pre className={s.pre}>{item.error ?? t.noError}</pre>
      <h3 className={s.detailHeading}>{t.userAgent}</h3>
      <pre className={s.pre}>{item.user_agent ?? t.noUserAgent}</pre>
      <h3 className={s.detailHeading}>{t.ip}</h3>
      <pre className={s.pre}>{item.ip ?? t.noIp}</pre>
    </article>
  );
}

export function AdminRequests() {
  const layout = useLayout();
  const token = useAuthStore((st) => st.jwtToken);
  const ids = useId();
  const statusErrorId = `${ids}-status`;
  const userErrorId = `${ids}-user`;
  const rangeErrorId = `${ids}-range`;

  const [route, setRoute] = useState("");
  const [period, setPeriod] = useState<Period>("24h");
  const [initialRange] = useState(() => presetRange("24h"));
  const [customFrom, setCustomFrom] = useState(() => toLocalInput(initialRange.from));
  const [customTo, setCustomTo] = useState(() => toLocalInput(initialRange.to));
  const [statusInput, setStatusInput] = useState("");
  const [username, setUsername] = useState("");
  const [userId, setUserId] = useState("");
  const [service, setService] = useState<Service>("");
  const [archived, setArchived] = useState(false);
  const [limit, setLimit] = useState<number>(PAGE_SIZE_OPTIONS[0]);
  const [errors, setErrors] = useState<{ status?: string; userId?: string; range?: string }>({});

  const [applied, setApplied] = useState<Applied>(emptyApplied);
  const [page, setPage] = useState(1);
  const [reloadKey, setReloadKey] = useState(0);
  const [view, setView] = useState<View>({ kind: "loading" });
  const [selected, setSelected] = useState<AdminRequestLog | null>(null);
  const [tab, setTab] = useState<Tab>("requests");
  const [novos, setNovos] = useState(0);
  const requestsTabRef = useRef<HTMLButtonElement>(null);
  const bugsTabRef = useRef<HTMLButtonElement>(null);
  const requestsTabId = `${ids}-tab-requests`;
  const requestsPanelId = `${ids}-panel-requests`;
  const bugsTabId = `${ids}-tab-bugs`;
  const bugsPanelId = `${ids}-panel-bugs`;

  const selectTab = (next: Tab) => {
    setTab(next);
    setSelected(null);
  };

  const onTabKeyDown = (event: KeyboardEvent<HTMLButtonElement>) => {
    if (
      event.key !== "ArrowLeft" &&
      event.key !== "ArrowRight" &&
      event.key !== "Home" &&
      event.key !== "End"
    ) {
      return;
    }
    event.preventDefault();
    const next: Tab =
      event.key === "Home" ? "requests" : event.key === "End" ? "bugs" : tab === "requests" ? "bugs" : "requests";
    selectTab(next);
    (next === "requests" ? requestsTabRef : bugsTabRef).current?.focus();
  };

  useEffect(() => {
    if (!token) {
      setView({ kind: "unauthorized" });
      return;
    }

    const controller = new AbortController();
    setView({ kind: "loading" });

    listRequestLogs(
      token,
      {
        from: applied.from,
        to: applied.to,
        route: applied.route || undefined,
        status: applied.status || undefined,
        username: applied.username || undefined,
        userId: applied.userId || undefined,
        service: applied.service || undefined,
        archived: applied.archived,
        page,
        limit: applied.limit,
      },
      controller.signal,
    )
      .then((data) => setView({ kind: "ready", data }))
      .catch((error: unknown) => {
        if (controller.signal.aborted) return;
        if (error instanceof AdminApiError) {
          if (error.status === 403) return setView({ kind: "forbidden" });
          if (error.status === 401) return setView({ kind: "unauthorized" });
          return setView({ kind: "error", message: error.message });
        }
        if ((error as { name?: string } | null)?.name === "AbortError") return;
        setView({ kind: "error", message: t.unexpectedError });
      });

    return () => controller.abort();
  }, [token, applied, page, reloadKey]);

  const apply = () => {
    const next: typeof errors = {};
    const status = parseStatusInput(statusInput);
    if (status.error) next.status = status.error;

    const trimmedUserId = userId.trim();
    if (trimmedUserId && !isUuid(trimmedUserId)) next.userId = "Informe um UUID válido.";

    const range = period === "custom" ? customRange(customFrom, customTo, archived) : { range: presetRange(period) };
    if (range.error || !range.range) next.range = range.error ?? "Intervalo inválido.";

    const resolvedRange = range.range;
    if (status.error || !resolvedRange || Object.keys(next).length > 0) {
      setErrors(next);
      return;
    }

    setErrors({});
    setApplied({
      from: resolvedRange.from,
      to: resolvedRange.to,
      route: route.trim().slice(0, 64),
      status: status.value ?? "",
      username: username.trim().slice(0, 64),
      userId: trimmedUserId,
      service,
      archived,
      limit: clampLimit(limit),
    });
    setPage(1);
    setSelected(null);
  };

  const reset = () => {
    const range = presetRange("24h");
    setRoute("");
    setPeriod("24h");
    setCustomFrom(toLocalInput(range.from));
    setCustomTo(toLocalInput(range.to));
    setStatusInput("");
    setUsername("");
    setUserId("");
    setService("");
    setArchived(false);
    setLimit(PAGE_SIZE_OPTIONS[0]);
    setErrors({});
    setApplied(emptyApplied());
    setPage(1);
    setSelected(null);
  };

  const totalPages = view.kind === "ready" ? pageCount(view.data.total, view.data.limit) : 1;
  const busy = view.kind === "loading";

  return (
    <main className={cx(s.screen, layout === "desktop" && s.desk)}>
      <header className={s.head}>
        <a className={s.back} href="#">
          {t.back}
        </a>
        <h1 className={s.title}>{tab === "requests" ? t.title : t.bugsTitle}</h1>
        <p className={s.lead}>{tab === "requests" ? t.lead : t.bugsLead}</p>
      </header>

      <div role="tablist" aria-label={t.tabs} className={s.tabs}>
        <button
          ref={requestsTabRef}
          type="button"
          role="tab"
          id={requestsTabId}
          aria-selected={tab === "requests"}
          aria-controls={requestsPanelId}
          tabIndex={tab === "requests" ? 0 : -1}
          className={s.tab}
          onClick={() => selectTab("requests")}
          onKeyDown={onTabKeyDown}
        >
          {t.tabRequests}
        </button>
        <button
          ref={bugsTabRef}
          type="button"
          role="tab"
          id={bugsTabId}
          aria-selected={tab === "bugs"}
          aria-controls={bugsPanelId}
          aria-label={novos > 0 ? t.bugsWithNew(novos) : t.tabBugs}
          tabIndex={tab === "bugs" ? 0 : -1}
          className={s.tab}
          onClick={() => selectTab("bugs")}
          onKeyDown={onTabKeyDown}
        >
          {t.tabBugs}
          {novos > 0 && (
            <span aria-hidden="true" className={s.count}>
              {novos}
            </span>
          )}
        </button>
      </div>

      <div
        role="tabpanel"
        id={requestsPanelId}
        aria-labelledby={requestsTabId}
        className={s.panel}
        hidden={tab !== "requests"}
        tabIndex={0}
      >
      <section aria-label={t.filters} className={s.filters}>
        <div className={s.fieldWide}>
          <span aria-hidden="true" className={cx("hw-rotulo", s.caption)}>
            {t.route}
          </span>
          <TextField
            label={t.route}
            placeholder={t.routePlaceholder}
            value={route}
            onChange={setRoute}
            onSubmit={apply}
            maxLength={64}
            autoCapitalize="none"
          />
        </div>

        <label className={s.field}>
          <span className={cx("hw-rotulo", s.caption)}>{t.period}</span>
          <select className={s.control} value={period} onChange={(e) => setPeriod(e.target.value as Period)}>
            {PERIODS.map((p) => (
              <option key={p.value} value={p.value}>
                {p.label}
              </option>
            ))}
          </select>
        </label>

        {period === "custom" && (
          <>
            <label className={s.field}>
              <span className={cx("hw-rotulo", s.caption)}>{t.from}</span>
              <input
                type="datetime-local"
                className={s.control}
                value={customFrom}
                onChange={(e) => setCustomFrom(e.target.value)}
                aria-invalid={errors.range ? true : undefined}
                aria-describedby={errors.range ? rangeErrorId : undefined}
              />
            </label>
            <label className={s.field}>
              <span className={cx("hw-rotulo", s.caption)}>{t.to}</span>
              <input
                type="datetime-local"
                className={s.control}
                value={customTo}
                onChange={(e) => setCustomTo(e.target.value)}
                aria-invalid={errors.range ? true : undefined}
                aria-describedby={errors.range ? rangeErrorId : undefined}
              />
            </label>
          </>
        )}

        <div className={s.field}>
          <span aria-hidden="true" className={cx("hw-rotulo", s.caption)}>
            {t.status}
          </span>
          <TextField
            label={t.status}
            placeholder={t.statusPlaceholder}
            value={statusInput}
            onChange={setStatusInput}
            onSubmit={apply}
            maxLength={8}
            autoCapitalize="none"
            aria-invalid={errors.status ? true : undefined}
            aria-describedby={errors.status ? statusErrorId : undefined}
          />
        </div>

        <div className={s.field}>
          <span aria-hidden="true" className={cx("hw-rotulo", s.caption)}>
            {t.username}
          </span>
          <TextField
            label={t.username}
            placeholder={t.usernamePlaceholder}
            value={username}
            onChange={setUsername}
            onSubmit={apply}
            maxLength={64}
          />
        </div>

        <div className={s.field}>
          <span aria-hidden="true" className={cx("hw-rotulo", s.caption)}>
            {t.userId}
          </span>
          <TextField
            label={t.userId}
            placeholder={t.userIdPlaceholder}
            value={userId}
            onChange={setUserId}
            onSubmit={apply}
            maxLength={36}
            autoCapitalize="none"
            aria-invalid={errors.userId ? true : undefined}
            aria-describedby={errors.userId ? userErrorId : undefined}
          />
        </div>

        <label className={s.field}>
          <span className={cx("hw-rotulo", s.caption)}>{t.service}</span>
          <select className={s.control} value={service} onChange={(e) => setService(e.target.value as Service)}>
            <option value="">{t.allServices}</option>
            <option value="backend">backend</option>
            <option value="gameServer">gameServer</option>
          </select>
        </label>

        <label className={s.field}>
          <span className={cx("hw-rotulo", s.caption)}>{t.limit}</span>
          <select className={s.control} value={limit} onChange={(e) => setLimit(Number(e.target.value))}>
            {PAGE_SIZE_OPTIONS.map((n) => (
              <option key={n} value={n}>
                {n}
              </option>
            ))}
          </select>
        </label>

        <div className={s.toggleField}>
          <Toggle label={t.archived} hint={t.archivedHint} value={archived} onChange={setArchived} />
        </div>

        <div className={s.actions}>
          <Button variant="outline" onClick={apply}>
            {t.apply}
          </Button>
          <Button variant="link" onClick={reset}>
            {t.reset}
          </Button>
        </div>

        {errors.status && (
          <p id={statusErrorId} role="alert" className={s.errorText}>
            {errors.status}
          </p>
        )}
        {errors.userId && (
          <p id={userErrorId} role="alert" className={s.errorText}>
            {errors.userId}
          </p>
        )}
        {errors.range && (
          <p id={rangeErrorId} role="alert" className={s.errorText}>
            {errors.range}
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
              <span>{t.pageLabel(page, totalPages)}</span>
            </div>

            <div className={s.tableWrap}>
              <table className={s.table}>
                <caption className={s.srOnly}>{t.tableCaption}</caption>
                <thead>
                  <tr>
                    <th scope="col">{t.at}</th>
                    <th scope="col">{t.serviceField}</th>
                    <th scope="col">{t.routeField}</th>
                    <th scope="col">{t.statusField}</th>
                    <th scope="col">{t.duration}</th>
                    <th scope="col">{t.usernameField}</th>
                    <th scope="col">{t.ip}</th>
                    <th scope="col">
                      <span className={s.srOnly}>{t.details}</span>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {view.data.items.map((item) => (
                    <tr key={item.id} className={cx(item.status != null && item.status >= 400 && s.rowError)}>
                      <td className={s.nowrap}>{formatAt(item.at)}</td>
                      <td>{item.service}</td>
                      <td>
                        <span className={s.method}>{item.method ?? "-"}</span>
                        <span className={s.route}>{item.route ?? item.path ?? "-"}</span>
                      </td>
                      <td>
                        <span className={cx(s.status, statusClass(item.status))}>{item.status ?? "-"}</span>
                      </td>
                      <td className={s.nowrap}>{item.duration_ms == null ? "-" : `${item.duration_ms} ms`}</td>
                      <td>{item.username ?? item.user_id ?? "-"}</td>
                      <td>{item.ip ?? "-"}</td>
                      <td>
                        <Button
                          variant="link"
                          className={s.details}
                          aria-label={t.detailsOf(item.id)}
                          onClick={() => setSelected(item)}
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
      </div>

      <div
        role="tabpanel"
        id={bugsPanelId}
        aria-labelledby={bugsTabId}
        className={s.panel}
        hidden={tab !== "bugs"}
        tabIndex={0}
      >
        <AdminBugs active={tab === "bugs"} onNovosChange={setNovos} />
      </div>

      <Sheet
        open={!!selected && tab === "requests"}
        onClose={() => setSelected(null)}
        label={selected ? t.detailTitle(selected.id) : t.details}
        mode={layout}
        backLabel={t.close}
        fill
      >
        {selected && <Detail item={selected} />}
      </Sheet>
    </main>
  );
}
