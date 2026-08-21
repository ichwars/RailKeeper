import { useRef } from "react";
import { X } from "lucide-react";

import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";

export function LocomotiveComparisonDialog({ item, onClose }: {
  item: DigitalCenterWorkItem;
  onClose: () => void;
}) {
  const closeButtonRef = useRef<HTMLButtonElement>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(onClose, closeButtonRef);
  const rows = [
    ["Name", item.center.name ?? "–", item.railkeeper.name ?? "–"],
    ["Decoder-Adresse", item.center.decoderAddress ?? "–", item.railkeeper.decoderAddress ?? "–"],
    ["Protokoll", item.center.protocol ?? "–", item.railkeeper.protocol ?? "–"]
  ];
  return (
    <>
      <span ref={anchorRef} aria-hidden="true" />
      <div ref={layerRef} className="digital-comparison-layer" role="dialog" aria-modal="true"
        aria-label={`Lok-Vergleich ${item.name}`} onKeyDown={onKeyDown}>
        <section className="digital-comparison-dialog">
          <header><div><p className="eyebrow">LOK-ABGLEICH</p><h2>{item.name}</h2></div>
            <button ref={closeButtonRef} type="button" className="digital-center-icon-button"
              aria-label="Vergleich schließen" onClick={onClose}>
              <X size={19} aria-hidden="true" />
            </button></header>
          <table><thead><tr><th>Feld</th><th>Digitalzentrale</th><th>RailKeeper</th></tr></thead>
            <tbody>{rows.map(([label, center, railkeeper]) => <tr key={label}><th>{label}</th>
              <td title={String(center)}>{center}</td>
              <td title={String(railkeeper)}>{railkeeper}</td></tr>)}</tbody></table>
          <footer><button type="button" className="digital-center-button"
            onClick={onClose}>Schließen</button></footer>
        </section>
      </div>
    </>
  );
}
