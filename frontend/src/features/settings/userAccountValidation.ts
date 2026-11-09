import type { SettingsAuthController } from "./useSettingsAuth";
import type { useI18n } from "../../shared/i18n";

export function userAccountValidation(
  userForm: SettingsAuthController["userForm"],
  editingUser: SettingsAuthController["editingUser"],
  userSaving: boolean,
  t: ReturnType<typeof useI18n>["t"]
) {
  const passwordTouched = userForm.password.length > 0 || userForm.confirmPassword.length > 0;
  const passwordRequired = !editingUser;
  const userPasswordTooShort = (passwordRequired || passwordTouched) && userForm.password.length < 12;
  const userPasswordMismatch = passwordTouched && userForm.password !== userForm.confirmPassword;
  const userPasswordMessage = passwordTouched && userPasswordTooShort
    ? t("settings.users.passwordTooShort")
    : userPasswordMismatch
      ? t("setup.passwordMismatch")
      : "";
  const canSaveUser =
    !userSaving &&
    userForm.username.trim().length > 0 &&
    userForm.roles.length > 0 &&
    !userPasswordTooShort &&
    !userPasswordMismatch;
  return { canSaveUser, userPasswordMessage, passwordRequired, passwordTouched };
}
