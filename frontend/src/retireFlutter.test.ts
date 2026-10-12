// @vitest-environment jsdom
import { afterEach, expect, it, vi } from "vitest";
import { retireFlutter } from "./retireFlutter";
afterEach(() => {
  vi.unstubAllGlobals();
});
it("retires only the Flutter worker belonging to this app", async () => {
  const scope = new URL(import.meta.env.BASE_URL, location.origin).href;
  const unregister = vi.fn(),
    other = vi.fn();
  Object.defineProperty(navigator, "serviceWorker", {
    configurable: true,
    value: {
      getRegistrations: async () => [
        {
          scope,
          active: { scriptURL: scope + "flutter_service_worker.js" },
          unregister,
        },
        {
          scope: scope + "other/",
          active: { scriptURL: scope + "other/worker.js" },
          unregister: other,
        },
      ],
    },
  });
  const remove = vi.fn();
  vi.stubGlobal("caches", { delete: remove });
  await retireFlutter();
  expect(unregister).toHaveBeenCalledTimes(1);
  expect(other).not.toHaveBeenCalled();
  expect(remove.mock.calls.map(([name]) => name)).toEqual([
    "flutter-app-cache",
    "flutter-temp-cache",
    "flutter-app-manifest",
  ]);
});
