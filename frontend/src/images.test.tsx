// @vitest-environment jsdom
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
vi.mock("./api", () => ({ request: vi.fn() }));
import { request } from "./api";
import { clearImages, useImage } from "./images";
beforeEach(() => {
  URL.createObjectURL = vi.fn(() => "blob:test");
  URL.revokeObjectURL = vi.fn();
});
afterEach(() => {
  cleanup();
  clearImages();
  vi.resetAllMocks();
});
it("shares a load and aborts it only after the last consumer leaves", () => {
  vi.mocked(request).mockImplementation(() => new Promise(() => {}));
  const first = renderHook(() => useImage("/db/mini/a"));
  const second = renderHook(() => useImage("/db/mini/a"));
  expect(request).toHaveBeenCalledTimes(1);
  const signal = vi.mocked(request).mock.calls[0][1]!;
  first.unmount();
  expect(signal.aborted).toBe(false);
  second.unmount();
  expect(signal.aborted).toBe(true);
});
it("revokes cached image URLs when the account changes", async () => {
  vi.mocked(request).mockResolvedValue({
    blob: async () => new Blob(["jpg"], { type: "image/jpeg" }),
  } as Response);
  const { result } = renderHook(() => useImage("/db/mini/b"));
  await waitFor(() => expect(result.current.url).toBe("blob:test"));
  clearImages();
  expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:test");
});
it("does not expose a pending response after a session is cleared", async () => {
  let resolve!: (r: Response) => void;
  vi.mocked(request).mockImplementation(
    () =>
      new Promise((r) => {
        resolve = r;
      }),
  );
  const { result } = renderHook(() => useImage("/db/mini/c"));
  clearImages();
  await act(async () =>
    resolve({
      blob: async () => new Blob(["jpg"], { type: "image/jpeg" }),
    } as Response),
  );
  expect(result.current.url).toBeUndefined();
  expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:test");
});
