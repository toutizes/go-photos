import { expect, it } from "vitest";
import { signInErrorMessage } from "./authErrors";
it("distinguishes configuration failures from popup blocking", () => {
  expect(signInErrorMessage({ code: "auth/unauthorized-domain" })).toContain(
    "domaine n’est pas autorisé",
  );
  expect(signInErrorMessage({ code: "auth/popup-blocked" })).toContain(
    "a bloqué la fenêtre",
  );
  expect(signInErrorMessage({ code: "auth/internal-error" })).toContain(
    "auth/internal-error",
  );
});
it("does not expose raw error messages or arbitrary data", () => {
  expect(
    signInErrorMessage({ message: "secret", code: "secret" }),
  ).not.toContain("secret");
  expect(signInErrorMessage(null)).toContain("a échoué");
});
