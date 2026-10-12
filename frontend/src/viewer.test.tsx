// @vitest-environment jsdom
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { Viewer } from "./Viewer";
import { Grid } from "./Grid";
import type { Photo } from "./model";
vi.mock("./images", () => ({
  useImage: (path?: string) => ({
    url: path ? "data:image/png;base64,test" : undefined,
    retry: vi.fn(),
  }),
  PhotoImage: ({ alt }: { alt: string }) => <span>{alt}</span>,
}));
const photos: Photo[] = [1, 2, 3].map((Id) => ({
  Id,
  Ad: "album",
  In: `photo${Id}.jpg`,
  Its: 0,
  Fts: 0,
  W: 1600,
  H: 800,
  Kwd: ["famille"],
}));
beforeEach(() => {
  vi.stubGlobal(
    "ResizeObserver",
    class {
      constructor(private cb: ResizeObserverCallback) {}
      observe() {
        this.cb(
          [
            { contentRect: { width: 800, height: 400 } },
          ] as ResizeObserverEntry[],
          this as unknown as ResizeObserver,
        );
      }
      disconnect() {}
    },
  );
  vi.spyOn(HTMLElement.prototype, "getBoundingClientRect").mockReturnValue({
    width: 800,
    height: 400,
    left: 0,
    top: 0,
    right: 800,
    bottom: 400,
    x: 0,
    y: 0,
    toJSON: () => ({}),
  });
  HTMLElement.prototype.setPointerCapture = () => {};
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});
function gesture(el: Element, type: string, x: number, y: number, id = 1) {
  const e = new Event(type, { bubbles: true });
  Object.assign(e, {
    clientX: x,
    clientY: y,
    pointerId: id,
    pointerType: "touch",
    button: 0,
  });
  fireEvent(el, e);
}
it("supports swiping at fit scale and panning without navigating when zoomed", () => {
  const onStep = vi.fn();
  const { container } = render(
    <Viewer
      photos={photos}
      id={2}
      full
      onStep={onStep}
      onFull={() => {}}
      onClose={() => {}}
      onSearch={() => {}}
    />,
  );
  const stage = container.querySelector(".image-stage")!;
  gesture(stage, "pointerdown", 600, 200);
  gesture(stage, "pointermove", 400, 200);
  gesture(stage, "pointerup", 400, 200);
  expect(onStep).toHaveBeenCalledWith(3);
  onStep.mockClear();
  fireEvent.click(screen.getByRole("button", { name: "Agrandir" }));
  gesture(stage, "pointerdown", 600, 200);
  gesture(stage, "pointermove", 400, 200);
  gesture(stage, "pointerup", 400, 200);
  expect(onStep).not.toHaveBeenCalled();
  expect(
    container.querySelector(".viewer-image")?.getAttribute("style"),
  ).toContain("scale(1.5)");
  expect(screen.queryByText("famille")).toBeNull();
});
it("distinguishes a pinch from swipe navigation", () => {
  const onStep = vi.fn();
  const { container } = render(
    <Viewer
      photos={photos}
      id={2}
      full
      onStep={onStep}
      onFull={() => {}}
      onClose={() => {}}
      onSearch={() => {}}
    />,
  );
  const stage = container.querySelector(".image-stage")!;
  gesture(stage, "pointerdown", 300, 200, 1);
  gesture(stage, "pointerdown", 500, 200, 2);
  gesture(stage, "pointermove", 600, 200, 2);
  gesture(stage, "pointerup", 600, 200, 2);
  gesture(stage, "pointerup", 300, 200, 1);
  expect(
    container.querySelector(".viewer-image")?.getAttribute("style"),
  ).toContain("scale(1.5)");
  expect(onStep).not.toHaveBeenCalled();
});
it("double tap opens fullscreen from metadata", () => {
  const onFull = vi.fn();
  const { container } = render(
    <Viewer
      photos={photos}
      id={2}
      full={false}
      onStep={() => {}}
      onFull={onFull}
      onClose={() => {}}
      onSearch={() => {}}
    />,
  );
  const stage = container.querySelector(".image-stage")!;
  for (let i = 0; i < 2; i++) {
    gesture(stage, "pointerdown", 400, 200);
    gesture(stage, "pointerup", 400, 200);
  }
  expect(onFull).toHaveBeenCalledTimes(1);
  expect(screen.getByRole("button", { name: "famille" })).toBeDefined();
});
it("does not create 70,000 image elements and restores a distant selected row", () => {
  const many = Array.from({ length: 70000 }, (_, i) => ({
    ...photos[0],
    Id: i + 1,
    In: `photo${i + 1}.jpg`,
  }));
  const { container, rerender } = render(
    <Grid photos={many} open={() => {}} hidden={false} />,
  );
  expect(container.querySelectorAll(".photo-tile").length).toBeLessThan(50);
  rerender(<Grid photos={many} open={() => {}} hidden activeId={60000} />);
  rerender(
    <Grid photos={many} open={() => {}} hidden={false} activeId={60000} />,
  );
  expect(container.querySelectorAll(".photo-tile").length).toBeLessThan(50);
  expect(
    screen.getByRole("button", { name: "Ouvrir photo60000.jpg" }),
  ).toBeDefined();
});
