export interface Stereo {
  Dx: number;
  Dy: number;
  AnaDx?: number;
  AnaDy?: number;
}
export interface Photo {
  Id: number;
  Ad: string;
  In: string;
  Its: number;
  Fts: number;
  H: number;
  W: number;
  Kwd: string[] | null;
  Stereo?: Stereo;
}
export interface Album {
  Id: string;
  Ats: number;
  Dts: number;
  Nimgs: number;
  Cov: number;
  CovName: string;
  PreviewIds: number[] | null;
}
export interface ActivityGroup {
  keywords: { keyword: string; count: number }[];
  recent_images: { id: number; name: string }[];
  total_count: number;
}
export interface QueryEntry {
  query: string;
  timestamp: string;
  kind: string;
}
export const photoPath = (p: Photo, size: "mini" | "midi" | "maxi" = "midi") =>
  `/db/${size}/${p.Ad.split("/").map(encodeURIComponent).join("/")}/${encodeURIComponent(p.In)}?v=${p.Fts}`;
export const montagePath = (ids: number[]) =>
  `/db/montage/360x360-${ids.join("-")}`;
export const stereoOf = (p: Photo): Stereo | undefined =>
  p.Stereo ??
  ((p.Kwd ?? []).some((k) => /^stereo:|^(stereo|stéreo|stéréo)$/.test(k))
    ? { Dx: 0, Dy: 0 }
    : undefined);
export const todayQuery = () => {
  const d = new Date();
  return `${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
};
export const dateLabel = (seconds: number) =>
  new Intl.DateTimeFormat("fr", { dateStyle: "long" }).format(
    new Date(seconds * 1000),
  );
export function sortedPhotos(photos: Photo[], query: string) {
  const album = query
    .split(",")
    .some((term) => term.trim().startsWith("album:"));
  return [...photos].sort(
    (a, b) =>
      (album ? a.Its - b.Its : b.Its - a.Its) ||
      b.In.localeCompare(a.In) ||
      a.Id - b.Id,
  );
}
export interface Route {
  page: "activity" | "albums" | "photos" | "admin";
  q: string;
  id?: number;
  full: boolean;
}
export function readRoute(hash = location.hash): Route {
  const [path, search = ""] = hash.replace(/^#\/?/, "").split("?");
  const params = new URLSearchParams(search);
  const id =
    params.has("id") && /^\d+$/.test(params.get("id")!)
      ? Number(params.get("id"))
      : undefined;
  return {
    page: ["activity", "albums", "photos", "admin"].includes(path)
      ? (path as Route["page"])
      : "activity",
    q: params.get("q") ?? "",
    id,
    full: path === "photos" && id !== undefined && params.get("full") === "1",
  };
}
export function routeHash(r: Route) {
  const p = new URLSearchParams();
  if (r.q) p.set("q", r.q);
  if (r.id !== undefined) p.set("id", String(r.id));
  if (r.full && r.id !== undefined) p.set("full", "1");
  return `#/${r.page}${p.size ? `?${p}` : ""}`;
}
