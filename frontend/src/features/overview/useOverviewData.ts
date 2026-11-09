import { useCallback, useEffect, useRef, useState } from "react";
import {
  api,
  type AccessoryArticleListResult,
  type DigitalCenterSettings,
  type ECoSLiveStatus,
  type OverviewValuation,
  type Vehicle
} from "../../shared/api";
import { observeAction } from "../../shared/asyncAction";
import type { Language } from "../../shared/i18n";
import { translate } from "../../shared/i18n";



type useOverviewDataProps = {
  canManageDigital: boolean;
  language: Language;
};

export function useOverviewData({
  canManageDigital, language,
}: useOverviewDataProps) {
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);

  const [accessories, setAccessories] = useState<AccessoryArticleListResult | null>(null);

  const [valuation, setValuation] = useState<OverviewValuation | null>(null);

  const [digitalSettings, setDigitalSettings] = useState<DigitalCenterSettings | null>(null);

  const [liveStatus, setLiveStatus] = useState<ECoSLiveStatus | null>(null);

  const [vehicleLoading, setVehicleLoading] = useState(true);

  const [accessoryLoading, setAccessoryLoading] = useState(true);

  const [valuationLoading, setValuationLoading] = useState(true);

  const [digitalLoading, setDigitalLoading] = useState(canManageDigital);

  const [vehicleError, setVehicleError] = useState("");

  const [accessoryError, setAccessoryError] = useState("");

  const [valuationError, setValuationError] = useState("");

  const [digitalError, setDigitalError] = useState("");

  const valuationRequestID = useRef(0);

  const loadVehicles = useCallback(() => {
    setVehicleLoading(true);
    setVehicleError("");
    return api.vehicles().then(setVehicles).catch((error: Error) => setVehicleError(error.message))
      .finally(() => setVehicleLoading(false));
  }, []);

  const loadAccessories = useCallback(() => {
    setAccessoryLoading(true);
    setAccessoryError("");
    return api.accessoryArticles().then(setAccessories)
      .catch((error: Error) => setAccessoryError(error.message))
      .finally(() => setAccessoryLoading(false));
  }, []);

  const loadValuation = useCallback(() => {
    setValuationLoading(true);
    setValuationError("");
    const requestID = valuationRequestID.current + 1;
    valuationRequestID.current = requestID;
    return api.overviewValuation().then((nextValuation) => {
      if (requestID === valuationRequestID.current) setValuation(nextValuation);
    }).catch(() => {
      if (requestID === valuationRequestID.current) {
        setValuationError(translate(language, "overview.valuation.error"));
      }
    }).finally(() => {
      if (requestID === valuationRequestID.current) setValuationLoading(false);
    });
  }, [language]);

  const loadDigital = useCallback(() => {
    if (!canManageDigital) return Promise.resolve();
    setDigitalLoading(true);
    setDigitalError("");
    return api.digitalSettings().then(async (settings) => {
      setDigitalSettings(settings);
      if (settings.provider === "ecos" && settings.ecos.enabled && settings.ecos.host.trim()) {
        try {
          setLiveStatus(await api.getECoSLiveStatus());
        } catch {
          setLiveStatus(null);
        }
      } else {
        setLiveStatus(null);
      }
    }).catch((error: Error) => setDigitalError(error.message))
      .finally(() => setDigitalLoading(false));
  }, [canManageDigital]);

  useEffect(() => {
    observeAction(loadVehicles());
    observeAction(loadAccessories());
  }, [loadAccessories, loadVehicles]);

  useEffect(() => {
    observeAction(loadValuation());
  }, [loadValuation]);

  useEffect(() => {
    observeAction(loadDigital());
  }, [loadDigital]);

  return {
    vehicles, accessories, valuation, digitalSettings, liveStatus, vehicleLoading, accessoryLoading,
    valuationLoading, digitalLoading, vehicleError, accessoryError, valuationError, digitalError,
    loadVehicles, loadAccessories, loadValuation, loadDigital,
  };
}
