import { useState } from "react";
import { Download, X } from "lucide-react";
import { download } from "./api";
export function Downloads({ query, count }: { query: string; count: number }) {
  const [open, setOpen] = useState(false),
    [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  async function save(original: boolean) {
    setBusy(true);
    setError("");
    try {
      await download(
        `/db/viewer?command=download&q=${encodeURIComponent(query)}&s=${original ? "O" : "M"}`,
        "toutizes-photos.zip",
      );
      setOpen(false);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="download-menu">
      <button
        aria-label="Télécharger les photos"
        disabled={!count}
        onClick={() => setOpen(!open)}
      >
        <Download size={17} />
        <span>Télécharger</span>
      </button>
      {open && (
        <div
          className="download-popover"
          role="dialog"
          aria-label="Télécharger les photos"
        >
          <button
            className="icon-button popover-close"
            aria-label="Fermer le téléchargement"
            onClick={() => setOpen(false)}
          >
            <X size={17} />
          </button>
          <h2>Télécharger les photos</h2>
          <p>Un fichier ZIP, limité à 100 photos par téléchargement.</p>
          {count > 100 && (
            <p>
              Cette recherche contient {count.toLocaleString("fr")} photos. Le
              serveur choisira les 100 photos incluses dans le ZIP.
            </p>
          )}
          <button disabled={busy} onClick={() => save(false)}>
            Format écran · plus léger
          </button>
          <button disabled={busy} onClick={() => save(true)}>
            Originaux · pleine résolution
          </button>
          {busy && <p role="status">Préparation du téléchargement…</p>}
          {error && <p role="alert">{error}</p>}
        </div>
      )}
    </div>
  );
}
