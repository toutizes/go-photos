// @vitest-environment jsdom
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, expect, it } from "vitest";
import { useRoute } from "./useRoute";
afterEach(cleanup);
it("keeps the current photo on close, then returns to the grid", async () => {
  history.replaceState({}, "", "#/photos?q=family");
  const { result } = renderHook(useRoute);
  act(() =>
    result.current.navigate({ ...result.current.route, id: 1 }, false, "photo"),
  );
  act(() =>
    result.current.navigate(
      { ...result.current.route, full: true },
      false,
      "full",
    ),
  );
  act(() =>
    result.current.navigate({ ...result.current.route, id: 2 }, true, "full"),
  );
  act(() => result.current.close());
  await waitFor(() =>
    expect(result.current.route).toMatchObject({ id: 2, full: false }),
  );
  act(() => result.current.close());
  await waitFor(() => expect(result.current.route.id).toBeUndefined());
});
it("reconciles the visible route with native browser Back and Forward", async () => {
  history.replaceState({}, "", "#/photos?q=family&id=1");
  const { result } = renderHook(useRoute);
  act(() => result.current.navigate({ ...result.current.route, id: 2 }));
  act(() => history.back());
  await waitFor(() => expect(result.current.route.id).toBe(1));
  act(() => history.forward());
  await waitFor(() => expect(result.current.route.id).toBe(2));
});
it("closes a directly opened fullscreen link without leaving the site", () => {
  history.replaceState({}, "", "#/photos?q=family&id=7&full=1");
  const { result } = renderHook(useRoute);
  act(() => result.current.close());
  expect(result.current.route).toMatchObject({ id: 7, full: false });
});
it("closing a directly opened fullscreen photo twice stays within the app", () => {
  history.replaceState({}, "", "#/photos?q=family&id=9&full=1");
  const { result } = renderHook(useRoute);
  act(() => result.current.close());
  act(() => result.current.close());
  expect(result.current.route).toMatchObject({
    page: "photos",
    q: "family",
    id: undefined,
    full: false,
  });
});
