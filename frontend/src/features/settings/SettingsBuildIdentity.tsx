import type { VersionInfo } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

export function SettingsBuildIdentity({ build }: { build?: VersionInfo["build"] }) {
  const { t } = useI18n();
  if (!build) return null;
  return <><br /><small title={build.commit}>
    {t("settings.updates.build.label")}: <code>{build.shortCommit || t("settings.updates.build.unknownCommit")}</code>
    {" · "}{t(`settings.updates.build.${build.status}`)}
  </small></>;
}
