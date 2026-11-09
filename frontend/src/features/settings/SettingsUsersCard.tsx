import { Pencil, Trash2, UserCog, Users } from "lucide-react";
import type { UserAccount } from "../../shared/api";
import { formatDateTime } from "./settingsModel";
import type { SettingsAuthController } from "./useSettingsAuth";

type SettingsUsersCardProps = {
  t: (key: string, values?: Record<string, string | number>) => string;
  userCountLabel: string;
  canManageUsers: SettingsAuthController["canManageUsers"];
  openUserCreateDialog: () => void;
  usersLoading: SettingsAuthController["usersLoading"];
  users: SettingsAuthController["users"];
  editingUser: SettingsAuthController["editingUser"];
  currentSession: SettingsAuthController["currentSession"];
  resetLoginBudget: SettingsAuthController["resetLoginBudget"];
  openUserEditDialog: (user: UserAccount) => void;
  deleteUser: SettingsAuthController["deleteUser"];
};

export function SettingsUsersCard({
  t, userCountLabel, canManageUsers, openUserCreateDialog, usersLoading, users, editingUser,
  currentSession, resetLoginBudget, openUserEditDialog, deleteUser,
}: SettingsUsersCardProps) {
  return (
    <section className="panel settings-card settings-tool-card user-management-card">
      <div className="settings-section-head">
        <div className="settings-card-title">
          <Users size={18} />
          <div>
            <h2 className="settings-title-with-pill">
              {t("settings.users.title")}
              <span className="settings-pill">{userCountLabel}</span>
            </h2>
            <p>{t("settings.users.subtitle")}</p>
          </div>
        </div>
        <div className="user-management-toolbar">
          {canManageUsers && (
            <button type="button" className="secondary-button" onClick={openUserCreateDialog}>
              <UserCog size={16} />
              {t("settings.users.new")}
            </button>
          )}
        </div>
      </div>

      {!canManageUsers ? (
        <div className="current-user-card">
          <strong>{t("settings.users.adminRequired")}</strong>
          <span>{t("settings.users.adminHelp")}</span>
        </div>
      ) : (
        <div className="user-management-grid">
          <div className="table-wrap settings-inline-table user-table">
            <table>
              <thead>
                <tr>
                  <th>{t("settings.users.username")}</th>
                  <th>{t("auth.email")}</th>
                  <th>{t("settings.users.roles")}</th>
                  <th>{t("settings.users.created")}</th>
                  <th>{t("settings.sessions.actions")}</th>
                </tr>
              </thead>
              <tbody>
                {usersLoading ? (
                  <tr><td colSpan={5} className="loading-cell">{t("settings.users.loading")}</td></tr>
                ) : users.length === 0 ? (
                  <tr><td colSpan={5} className="loading-cell">{t("settings.users.empty")}</td></tr>
                ) : (
                  users.map((user) => (
                    <tr key={user.id} className={editingUser?.id === user.id ? "selected-row" : ""}>
                      <td>
                        <div className="user-cell">
                          <strong>{user.username}</strong>
                          {currentSession?.username === user.username && <span className="settings-pill active">{t("settings.users.current")}</span>}
                        </div>
                      </td>
                      <td>{user.email || "-"}</td>
                      <td>
                        <div className="role-chip-row">
                          {user.roles.map((role) => <span className="settings-pill" key={role}>{role}</span>)}
                        </div>
                      </td>
                      <td>{formatDateTime(user.createdAt)}</td>
                      <td>
                        <div className="table-actions">
                          <button type="button" className="secondary-button" onClick={() => resetLoginBudget(user)}>
                            {t("settings.users.unlockLogin")}
                          </button>
                          <button type="button" className="icon-button" onClick={() => openUserEditDialog(user)} aria-label={t("vehicles.edit")} title={t("vehicles.edit")}>
                            <Pencil size={16} />
                          </button>
                          <button type="button" className="icon-button danger" onClick={() => deleteUser(user)} aria-label={t("vehicles.delete")} title={t("vehicles.delete")}>
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </section>
  );
}
