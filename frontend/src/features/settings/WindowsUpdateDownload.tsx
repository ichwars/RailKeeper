import { Download, ExternalLink } from "lucide-react";

import type { VersionInfo } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

type WindowsUpdateDownloadProps = {
  info: VersionInfo;
};

export function WindowsUpdateDownload({ info }: WindowsUpdateDownloadProps) {
  const { t } = useI18n();
  const windowsPackage = info.updateAvailable ? info.windowsPackage : undefined;

  return <>
    {windowsPackage && <div className="windows-update-download">
      <a className="primary-button windows-update-download-button"
        href={windowsPackage.url} rel="noreferrer">
        <Download size={16} aria-hidden="true" />
        <span>{t("settings.updates.download", { version: windowsPackage.version })}</span>
      </a>
      <p>{t("settings.updates.downloadHelp")}</p>
    </div>}
    {info.releaseUrl && <a className="settings-link-row"
      href={info.releaseUrl} target="_blank" rel="noreferrer">
      <ExternalLink size={15} aria-hidden="true" />
      {t("settings.updates.openRelease")}
    </a>}
  </>;
}
