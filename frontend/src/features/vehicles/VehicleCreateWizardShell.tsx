import { X } from "lucide-react";
import { useLayoutEffect, type FormEventHandler, type ReactNode } from "react";

import { useI18n } from "../../shared/i18n";
import type { VehicleCreateStep } from "./vehicleCreateWizardState";

type VehicleCreateWizardShellProps = {
  step: VehicleCreateStep;
  summaries: Record<VehicleCreateStep, string>;
  children: ReactNode;
  footer: ReactNode;
  onClose: () => void;
  onSubmit: FormEventHandler<HTMLFormElement>;
};

const steps: Array<{ value: VehicleCreateStep; labelKey: string }> = [
  { value: "basics", labelKey: "vehicles.wizard.step1" },
  { value: "article", labelKey: "vehicles.wizard.step2" },
  { value: "details", labelKey: "vehicles.wizard.step3" }
];

export function VehicleCreateWizardShell({
  step,
  summaries,
  children,
  footer,
  onClose,
  onSubmit
}: VehicleCreateWizardShellProps) {
  const { t } = useI18n();
  const currentIndex = steps.findIndex((item) => item.value === step);

  useLayoutEffect(() => {
    const resetHorizontalScroll = () => {
      document.documentElement.scrollLeft = 0;
      document.body.scrollLeft = 0;
    };
    resetHorizontalScroll();
    const frame = window.requestAnimationFrame(resetHorizontalScroll);
    return () => window.cancelAnimationFrame(frame);
  }, [step]);

  return (
    <form className="vehicle-modal vehicle-create-wizard" onSubmit={onSubmit}
      role="dialog" aria-modal="true" aria-label={t("vehicles.wizard.title")}>
      <header className="modal-head vehicle-create-head">
        <div>
          <h2>{t("vehicles.wizard.title")}</h2>
          <p>{t("vehicles.wizard.subtitle")}</p>
        </div>
        <button type="button" className="icon-button" onClick={onClose}
          aria-label={t("vehicles.close")} title={t("vehicles.close")}><X size={18} /></button>
      </header>

      <div className="vehicle-create-layout">
        <aside className="vehicle-wizard-rail">
          <ol className="vehicle-wizard-steps" aria-label={t("vehicles.wizard.progress")}>
            {steps.map((item, index) => (
              <li key={item.value} className={index === currentIndex ? "active" : index < currentIndex ? "done" : ""}
                aria-current={index === currentIndex ? "step" : undefined}>
                <span>{index + 1}</span>
                <span className="vehicle-wizard-step-copy">
                  <strong>{t(item.labelKey)}</strong>
                  <small>{summaries[item.value]}</small>
                </span>
              </li>
            ))}
          </ol>
        </aside>
        <main className="modal-body vehicle-wizard-body">{children}</main>
      </div>

      <footer className="modal-actions vehicle-wizard-actions">{footer}</footer>
    </form>
  );
}
