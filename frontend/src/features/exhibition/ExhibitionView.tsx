import { ExhibitionWorkspacePage } from "./ExhibitionWorkspacePage";

export function ExhibitionView({ roles }: { roles: string[] }) {
  return <ExhibitionWorkspacePage roles={roles} />;
}
