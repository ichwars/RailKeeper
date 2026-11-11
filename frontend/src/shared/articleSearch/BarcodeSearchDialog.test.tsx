import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { setLanguage } from "../i18n";
import { BarcodeSearchDialog } from "./BarcodeSearchDialog";

const decodeFromStream = vi.hoisted(() => vi.fn());
vi.mock("@zxing/browser", () => ({
  BrowserMultiFormatReader: class {
    decodeFromStream = decodeFromStream;
  }
}));

const getUserMedia = vi.fn();

beforeEach(() => {
  setLanguage("de");
  vi.stubGlobal("isSecureContext", true);
  vi.stubGlobal("navigator", { mediaDevices: { getUserMedia } });
  getUserMedia.mockReset();
  decodeFromStream.mockReset();
});

afterEach(() => vi.unstubAllGlobals());

function deferred<Value>() {
  let resolve: (value: Value) => void = () => undefined;
  const promise = new Promise<Value>(accept => { resolve = accept; });
  return { promise, resolve };
}

function cameraStream() {
  const stop = vi.fn();
  return { stop, stream: { getTracks: () => [{ stop }] } };
}

function startScanner() {
  fireEvent.click(screen.getByRole("button", { name: "EAN scannen" }));
}

it("releases camera permission that resolves after the dialog unmounts", async () => {
  const { stream, stop } = cameraStream();
  const permission = deferred<typeof stream>();
  getUserMedia.mockReturnValueOnce(permission.promise);
  const { unmount } = render(<BarcodeSearchDialog value="" onValueChange={vi.fn()} onClose={vi.fn()} onSubmit={vi.fn()} />);
  startScanner();
  await waitFor(() => expect(getUserMedia).toHaveBeenCalledOnce());
  unmount();
  await act(async () => permission.resolve(stream));
  expect(stop).toHaveBeenCalledOnce();
  expect(decodeFromStream).not.toHaveBeenCalled();
});

it("stops a late decoder without detaching the preview of a restarted scan", async () => {
  const first = cameraStream();
  const second = cameraStream();
  const firstControls = { stop: vi.fn() };
  const secondControls = { stop: vi.fn() };
  const decoder = deferred<typeof firstControls>();
  const onValueChange = vi.fn();
  getUserMedia.mockResolvedValueOnce(first.stream).mockResolvedValueOnce(second.stream);
  decodeFromStream.mockReturnValueOnce(decoder.promise).mockResolvedValueOnce(secondControls);
  render(<BarcodeSearchDialog value="" onValueChange={onValueChange} onClose={vi.fn()} onSubmit={vi.fn()} />);
  startScanner();
  await waitFor(() => expect(decodeFromStream).toHaveBeenCalledOnce());
  startScanner();
  await waitFor(() => expect(decodeFromStream).toHaveBeenCalledTimes(2));
  expect(first.stop).toHaveBeenCalledOnce();
  expect(decodeFromStream.mock.calls[0][1]).not.toBe(decodeFromStream.mock.calls[1][1]);

  await act(async () => {
    decodeFromStream.mock.calls[0][2]({ getText: () => "1234567890123" }, undefined, firstControls);
    decoder.resolve(firstControls);
  });
  expect(onValueChange).not.toHaveBeenCalled();
  expect(firstControls.stop).toHaveBeenCalledOnce();
  expect(secondControls.stop).not.toHaveBeenCalled();

  fireEvent.click(screen.getByRole("button", { name: "Kamera schließen" }));
  expect(secondControls.stop).toHaveBeenCalledOnce();
});
