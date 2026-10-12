// @vitest-environment jsdom
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
vi.mock("./auth", () => ({
  watchAccount: vi.fn(),
  login: vi.fn(),
  logout: vi.fn(),
  isAdmin: () => false,
}));
vi.mock("./api", () => ({
  cachedSearch: () => [],
  clearSearches: vi.fn(),
  search: vi.fn().mockResolvedValue([]),
}));
vi.mock("./images", () => ({ clearImages: vi.fn() }));
vi.mock("./Grid", () => ({ Grid: () => null }));
vi.mock("./Viewer", () => ({ Viewer: () => null }));
vi.mock("./Downloads", () => ({ Downloads: () => null }));
vi.mock("./Collections", () => ({
  ActivityPage: () => <p>Activité</p>,
  AlbumsPage: () => null,
  AdminPage: () => <p>Historique privé</p>,
}));
import { watchAccount } from "./auth";
import { search } from "./api";
import { App } from "./App";
beforeEach(() => {
  vi.clearAllMocks();
  history.replaceState({}, "", "#/photos?q=été");
  vi.mocked(watchAccount).mockImplementation((fn) => {
    fn({ email: "visitor@example.org", displayName: "Visiteur" });
    return () => {};
  });
});
afterEach(cleanup);
it("logs an explicit repeat submission of the current search", async () => {
  render(<App />);
  await waitFor(() => expect(search).toHaveBeenCalledTimes(1));
  fireEvent.click(screen.getByRole("button", { name: "Lancer la recherche" }));
  await waitFor(() => expect(search).toHaveBeenCalledTimes(2));
  expect(search).toHaveBeenLastCalledWith("été", expect.any(AbortSignal));
});
it("shows the Google gate without fetching photos when signed out", () => {
  vi.mocked(watchAccount).mockImplementation((fn) => {
    fn(null);
    return () => {};
  });
  render(<App />);
  expect(
    screen.getByRole("button", { name: "Se connecter avec Google" }),
  ).toBeDefined();
  expect(search).not.toHaveBeenCalled();
});
it("does not expose the owner history screen via a visitor deep link", () => {
  history.replaceState({}, "", "#/admin");
  render(<App />);
  expect(screen.getByText("Accès réservé au propriétaire")).toBeDefined();
  expect(screen.queryByText("Historique privé")).toBeNull();
});
