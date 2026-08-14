import { useEffect, useMemo, useState } from "react";
import { Cable, CircleAlert } from "lucide-react";

import { api, type LayoutConfiguration, type ModulePortAnalysis, type ModulePortIssue } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

const emptyAnalysis: ModulePortAnalysis = { connections: [], issues: [] };

function issueLabel(issue: ModulePortIssue) {
  return issue.unitNames.map((unitName, index) =>
    `${unitName} · ${issue.portNames[index] || issue.portIds[index] || "?"}`).join(" ↔ ");
}

function connectionLabel(connection: ModulePortAnalysis["connections"][number]) {
  return `${connection.unitAName} · ${connection.portAName} ↔ ${connection.unitBName} · ${connection.portBName}`;
}

export function LayoutConfigurationPortAnalysis({ configuration }: {
  configuration: LayoutConfiguration | null;
}) {
  const [analysis, setAnalysis] = useState<ModulePortAnalysis>(emptyAnalysis);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const { t } = useI18n();
  const open = useMemo(() => analysis.issues.filter((issue) => issue.code === "open_port"), [analysis]);
  const incompatible = useMemo(() =>
    analysis.issues.filter((issue) => issue.code === "incompatible_port"), [analysis]);

  useEffect(() => {
    let active = true;
    if (!configuration) {
      setAnalysis(emptyAnalysis);
      setError("");
      return () => { active = false; };
    }
    setLoading(true);
    setError("");
    api.layoutConfigurationPortAnalysis(configuration.id)
      .then((result) => active && setAnalysis(result))
      .catch((reason: Error) => active && setError(reason.message))
      .finally(() => active && setLoading(false));
    return () => { active = false; };
  }, [configuration]);

  return <section className="layout-port-analysis" aria-label={t("layouts.portAnalysis.title")}>
    <div className="panel-title"><Cable size={17} /><h4>{t("layouts.portAnalysis.title")}</h4></div>
    {!configuration ? <p className="layout-empty">{t("layouts.portAnalysis.noSelection")}</p> :
      loading ? <p className="layout-empty">{t("layouts.portAnalysis.loading")}</p> :
        error ? <p className="form-message">{error}</p> : <>
          <div className="layout-port-analysis-summary">
            <span>{t(analysis.connections.length === 1 ? "layouts.portAnalysis.connectionOne" :
              "layouts.portAnalysis.connectionMany", { count: analysis.connections.length })}</span>
            <span>{t(open.length === 1 ? "layouts.portAnalysis.openOne" : "layouts.portAnalysis.openMany",
              { count: open.length })}</span>
            <span>{t(incompatible.length === 1 ? "layouts.portAnalysis.incompatibleOne" :
              "layouts.portAnalysis.incompatibleMany", { count: incompatible.length })}</span>
          </div>
          {analysis.connections.length === 0 && analysis.issues.length === 0 ?
            <p className="layout-empty">{t("layouts.portAnalysis.empty")}</p> : null}
          {analysis.connections.length > 0 ? <ul className="layout-port-analysis-list">
            {analysis.connections.map((connection) => <li key={`${connection.unitAId}:${connection.portAId}:${connection.portBId}`}>
              <Cable size={14} /><span>{connectionLabel(connection)}</span></li>)}
          </ul> : null}
          {analysis.issues.length > 0 ? <ul className="layout-port-analysis-list issues">
            {analysis.issues.map((issue, index) => <li key={`${issue.code}:${issue.portIds.join(":")}:${index}`}>
              <CircleAlert size={14} /><span>{issueLabel(issue)}</span></li>)}
          </ul> : null}
        </>}
  </section>;
}
