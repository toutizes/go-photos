// @vitest-environment jsdom
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
vi.mock("./api", () => ({ json: vi.fn().mockResolvedValue([]) }));
vi.mock("./images", () => ({ useImage: () => ({}) }));
import { json } from "./api";
import { AlbumsPage } from "./Collections";
afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});
it("requests album results for a regular query and logs repeat submissions", async () => {
  const props = {
    query: "famille, vacances",
    searchVersion: 0,
    onSearch: vi.fn(),
  };
  const { rerender } = render(<AlbumsPage {...props} />);
  await waitFor(() =>
    expect(json).toHaveBeenCalledWith(
      "/db/q?q=famille%2C%20vacances&kind=album",
      expect.any(AbortSignal),
    ),
  );
  await screen.findByText("Aucun album trouvé pour « famille, vacances »");
  rerender(<AlbumsPage {...props} searchVersion={1} />);
  await waitFor(() => expect(json).toHaveBeenCalledTimes(2));
});
it("lists all albums with an empty search", async () => {
  render(<AlbumsPage query="" searchVersion={0} onSearch={vi.fn()} />);
  await waitFor(() =>
    expect(json).toHaveBeenCalledWith(
      "/db/q?q=albums%3A&kind=album",
      expect.any(AbortSignal),
    ),
  );
});
