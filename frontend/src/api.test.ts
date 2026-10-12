import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
vi.mock("./auth", () => ({ token: vi.fn() }));
import { token } from "./auth";
import { cachedSearch, clearSearches, json, search } from "./api";
const fetchMock = vi.fn();
beforeEach(() => {
  clearSearches();
  vi.stubGlobal("fetch", fetchMock);
  vi.mocked(token).mockResolvedValue("valid-token");
});
afterEach(() => {
  vi.unstubAllGlobals();
  vi.resetAllMocks();
});
describe("authenticated API", () => {
  it("does not issue a request without a signed-in account", async () => {
    vi.mocked(token).mockResolvedValue(null);
    await expect(json("/db/q")).rejects.toThrow("Connectez-vous");
    expect(fetchMock).not.toHaveBeenCalled();
  });
  it("refreshes a rejected token once and retries", async () => {
    fetchMock
      .mockResolvedValueOnce(new Response("", { status: 401 }))
      .mockResolvedValueOnce(Response.json([]));
    expect(await json("/db/q")).toEqual([]);
    expect(token).toHaveBeenNthCalledWith(1, false);
    expect(token).toHaveBeenNthCalledWith(2, true);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
  it("stops retrying an expired session", async () => {
    fetchMock.mockResolvedValue(new Response("", { status: 401 }));
    await expect(json("/db/q")).rejects.toThrow("session a expiré");
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
  it("records repeat searches through the server even when results are cached", async () => {
    fetchMock.mockImplementation(() => Promise.resolve(Response.json([])));
    await search("été, famille");
    expect(cachedSearch("été, famille")).toEqual([]);
    await search("été, famille");
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(fetchMock.mock.calls[1][0]).toBe(
      "/db/q?q=%C3%A9t%C3%A9%2C%20famille",
    );
    expect(fetchMock.mock.calls[1][1].cache).toBe("no-store");
  });
  it("bounds query caching and clears it between accounts", async () => {
    fetchMock.mockImplementation(() => Promise.resolve(Response.json([])));
    for (const q of ["1", "2", "3", "4"]) await search(q);
    expect(cachedSearch("1")).toBeUndefined();
    expect(cachedSearch("4")).toEqual([]);
    clearSearches();
    expect(cachedSearch("4")).toBeUndefined();
  });
});
