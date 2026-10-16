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
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});
function gesture(
  el: Element,
  type: string,
  x: number,
  y: number,
  id = 1,
  pointerType = "touch",
) {
  const e = new Event(type, { bubbles: true });
  Object.assign(e, {
    clientX: x,
    clientY: y,
    pointerId: id,
    pointerType,
    button: 0,
  });
  fireEvent(el, e);
}
it("supports swiping at fit scale and panning without navigating when zoomed", () => {
  vi.useFakeTimers();
  const onStep = vi.fn();
  const { container } = render(
    <Viewer
      photos={photos}
      id={2}
      onStep={onStep}
      onClose={() => {}}
      onSearch={() => {}}
    />,
  );
  const stage = container.querySelector(".image-stage")!;
  gesture(stage, "pointerdown", 600, 200);
  gesture(stage, "pointermove", 400, 200);
  expect(
    container.querySelector(".photo-slider")?.getAttribute("style"),
  ).toContain("translateX(-200px)");
  expect(container.querySelectorAll(".neighbor-slide")).toHaveLength(2);
  gesture(stage, "pointerup", 400, 200);
  expect(onStep).not.toHaveBeenCalled();
  expect(
    container.querySelector(".photo-slider")?.getAttribute("style"),
  ).toContain("translateX(-800px)");
  act(() => vi.advanceTimersByTime(220));
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
  expect(screen.getByRole("button", { name: "famille" })).toBeDefined();
});
it("distinguishes a pinch from swipe navigation", () => {
  const onStep = vi.fn();
  const { container } = render(
    <Viewer
      photos={photos}
      id={2}
      onStep={onStep}
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
it.each(["touch", "mouse"])(
  "%s taps toggle metadata without closing, including when zoomed",
  (pointerType) => {
    const onClose = vi.fn();
    const { container } = render(
      <Viewer
        photos={photos}
        id={2}
        onStep={() => {}}
        onClose={onClose}
        onSearch={() => {}}
      />,
    );
    const stage = container.querySelector(".image-stage")!;
    expect(container.querySelector(".fullscreen")).not.toBeNull();
    expect(screen.getByText("photo2.jpg")).toBeDefined();
    fireEvent.click(screen.getByRole("button", { name: "Agrandir" }));
    gesture(stage, "pointerdown", 400, 200, 1, pointerType);
    gesture(stage, "pointerup", 400, 200, 1, pointerType);
    expect(screen.queryByRole("button", { name: "famille" })).toBeNull();
    expect(screen.getAllByRole("button")).toHaveLength(1);
    expect(
      screen.getByRole("button", { name: "Revenir à la grille" }),
    ).toBeDefined();
    expect(container.querySelector(".counter")).toBeNull();
    expect(screen.queryByRole("button", { name: "Photo suivante" })).toBeNull();
    expect(
      screen.queryByRole("button", { name: "Photo précédente" }),
    ).toBeNull();
    gesture(stage, "pointerdown", 400, 200, 1, pointerType);
    gesture(stage, "pointerup", 400, 200, 1, pointerType);
    expect(screen.getByRole("button", { name: "famille" })).toBeDefined();
    expect(screen.getByRole("button", { name: "Agrandir" })).toBeDefined();
    expect(
      screen.getByRole("button", { name: "Télécharger l’original" }),
    ).toBeDefined();
    expect(container.querySelector(".counter")).not.toBeNull();
    expect(
      screen.getByRole("button", { name: "Photo suivante" }),
    ).toBeDefined();
    expect(
      screen.getByRole("button", { name: "Photo précédente" }),
    ).toBeDefined();
    expect(
      container.querySelector(".photo-metadata")?.closest(".photo-slider"),
    ).toBeNull();
    expect(onClose).not.toHaveBeenCalled();
  },
);
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

it("snaps back short drags, canceled gestures and swipes past the first photo", () => {
  vi.useFakeTimers();
  const onStep = vi.fn();
  const { container } = render(
    <Viewer
      photos={photos}
      id={1}
      onStep={onStep}
      onClose={() => {}}
      onSearch={() => {}}
    />,
  );
  const stage = container.querySelector(".image-stage")!;
  for (const [end, event] of [
    [380, "pointerup"],
    [200, "pointercancel"],
    [600, "pointerup"],
  ] as const) {
    gesture(stage, "pointerdown", 400, 200);
    gesture(stage, "pointermove", end, 200);
    gesture(stage, event, end, 200);
    expect(
      container.querySelector(".photo-slider")?.getAttribute("style"),
    ).toContain("translateX(0px)");
    act(() => vi.advanceTimersByTime(220));
    expect(onStep).not.toHaveBeenCalled();
  }
});
it("respects reduced motion and cancels pending navigation when closed", () => {
  vi.useFakeTimers();
  vi.stubGlobal("matchMedia", () => ({ matches: true }));
  const onStep = vi.fn();
  const { container, unmount } = render(
    <Viewer
      photos={photos}
      id={2}
      onStep={onStep}
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
  vi.stubGlobal("matchMedia", () => ({ matches: false }));
  gesture(stage, "pointerdown", 600, 200);
  gesture(stage, "pointermove", 400, 200);
  gesture(stage, "pointerup", 400, 200);
  unmount();
  act(() => vi.advanceTimersByTime(220));
  expect(onStep).not.toHaveBeenCalled();
});
it("keeps keyword searches and close available in the unified viewer", () => {
  const onSearch = vi.fn(),
    onClose = vi.fn();
  render(
    <Viewer
      photos={photos}
      id={2}
      onStep={() => {}}
      onClose={onClose}
      onSearch={onSearch}
    />,
  );
  fireEvent.click(screen.getByRole("button", { name: "famille" }));
  expect(onSearch).toHaveBeenCalledWith("famille");
  fireEvent.click(screen.getByRole("button", { name: "Revenir à la grille" }));
  expect(onClose).toHaveBeenCalledTimes(1);
});

it("pulls keywords open without searching and collapses them with the handle", () => {
  const onSearch = vi.fn();
  render(
    <Viewer
      photos={photos}
      id={2}
      onStep={() => {}}
      onClose={() => {}}
      onSearch={onSearch}
    />,
  );
  const pill = screen.getByRole("button", { name: "famille" });
  gesture(pill, "pointerdown", 400, 350);
  gesture(pill, "pointermove", 400, 200);
  gesture(pill, "pointerup", 400, 200);
  fireEvent.click(pill);
  expect(onSearch).not.toHaveBeenCalled();
  const handle = screen.getByRole("button", { name: "Réduire les mots-clés" });
  expect(handle.getAttribute("aria-expanded")).toBe("true");
  gesture(handle, "pointerdown", 400, 200);
  gesture(handle, "pointermove", 400, 350);
  gesture(handle, "pointerup", 400, 350);
  fireEvent.click(handle);
  expect(
    screen
      .getByRole("button", { name: "Développer les mots-clés" })
      .getAttribute("aria-expanded"),
  ).toBe("false");
});

it.each([
  ["Photo suivante", -800, 3],
  ["Photo précédente", 800, 1],
] as const)(
  "animates %s while zoomed and ignores repeated clicks during the slide",
  (name, translation, id) => {
    vi.useFakeTimers();
    const onStep = vi.fn();
    const { container } = render(
      <Viewer
        photos={photos}
        id={2}
        onStep={onStep}
        onClose={() => {}}
        onSearch={() => {}}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Agrandir" }));
    const arrow = screen.getByRole("button", { name });
    fireEvent.click(arrow);
    expect(
      container.querySelector(".photo-slider")?.getAttribute("style"),
    ).toContain(`translateX(${translation}px)`);
    expect(
      container
        .querySelector(".photo-slider")
        ?.classList.contains("is-settling"),
    ).toBe(true);
    expect(container.querySelectorAll(".neighbor-slide")).toHaveLength(2);
    expect(onStep).not.toHaveBeenCalled();
    fireEvent.click(arrow);
    act(() => vi.advanceTimersByTime(220));
    expect(onStep).toHaveBeenCalledExactlyOnceWith(id);
  },
);

it.each([
  [-1, 3],
  [1, 1],
])(
  "navigates only after pulling past the zoomed edge (%s)",
  (direction, target) => {
    vi.useFakeTimers();
    const onStep = vi.fn();
    const { container } = render(
      <Viewer
        photos={photos}
        id={2}
        onStep={onStep}
        onClose={() => {}}
        onSearch={() => {}}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Agrandir" }));
    const stage = container.querySelector(".image-stage")!;
    // At 1.5x, the 800px-wide fitted image has 200px of pan in either direction.
    gesture(stage, "pointerdown", 400, 200);
    gesture(stage, "pointermove", 400 + direction * 200, 200);
    gesture(stage, "pointerup", 400 + direction * 200, 200);
    expect(onStep).not.toHaveBeenCalled();
    gesture(stage, "pointerdown", 400, 200);
    gesture(stage, "pointermove", 400 + direction * 80, 200);
    expect(
      container.querySelector(".photo-slider")?.getAttribute("style"),
    ).toContain(`translateX(${direction * 80}px)`);
    gesture(stage, "pointerup", 400 + direction * 80, 200);
    expect(onStep).not.toHaveBeenCalled();
    act(() => vi.advanceTimersByTime(220));
    expect(onStep).toHaveBeenCalledExactlyOnceWith(target);
  },
);

it("allows reversing an edge pull back into a zoomed pan", () => {
  const onStep = vi.fn();
  const { container } = render(
    <Viewer
      photos={photos}
      id={2}
      onStep={onStep}
      onClose={() => {}}
      onSearch={() => {}}
    />,
  );
  fireEvent.click(screen.getByRole("button", { name: "Agrandir" }));
  const stage = container.querySelector(".image-stage")!;
  gesture(stage, "pointerdown", 400, 200);
  gesture(stage, "pointermove", 100, 200);
  gesture(stage, "pointermove", 300, 200);
  gesture(stage, "pointerup", 300, 200);
  expect(
    container.querySelector(".photo-slider")?.getAttribute("style"),
  ).toContain("translateX(0px)");
  expect(
    container.querySelector(".viewer-image")?.getAttribute("style"),
  ).toContain("translate(-100px,0px)");
  expect(onStep).not.toHaveBeenCalled();
});

it("preserves hidden and visible controls across photo changes", () => {
  const props = {
    photos,
    onStep: vi.fn(),
    onClose: vi.fn(),
    onSearch: vi.fn(),
  };
  const { container, rerender } = render(<Viewer {...props} id={1} />);
  const tap = () => {
    const stage = container.querySelector(".image-stage")!;
    gesture(stage, "pointerdown", 400, 200);
    gesture(stage, "pointerup", 400, 200);
  };
  tap();
  rerender(<Viewer {...props} id={2} />);
  expect(screen.getAllByRole("button")).toHaveLength(1);
  expect(
    screen.getByRole("button", { name: "Revenir à la grille" }),
  ).toBeDefined();
  tap();
  rerender(<Viewer {...props} id={3} />);
  expect(screen.getByRole("button", { name: "Agrandir" })).toBeDefined();
  expect(screen.getByRole("button", { name: "famille" })).toBeDefined();
});
