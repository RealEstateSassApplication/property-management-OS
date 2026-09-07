import { AppShell } from "../components/app-shell";
import {
  AdminConfigurationError,
  getOrganizationSettings,
  listOrganizationMembers,
  listPortalLinks,
  type OrganizationMember,
  type OrganizationSettings,
  type PortalLink,
} from "../lib/admin";
import { listOwners, listTenants, type Owner, type Tenant } from "../lib/property-os";
import {
  createMemberAction,
  linkOwnerPortalAction,
  linkTenantPortalAction,
  removeMemberAction,
  unlinkOwnerPortalAction,
  unlinkTenantPortalAction,
  updateMemberRoleAction,
  updateOrganizationAction,
} from "./actions";

export const dynamic = "force-dynamic";

const roles: OrganizationMember["role"][] = ["admin", "manager", "accountant", "maintenance", "viewer", "owner", "tenant", "agent"];

export default async function SettingsPage() {
  let settings: OrganizationSettings | null = null;
  let members: OrganizationMember[] = [];
  let links: PortalLink[] = [];
  let owners: Owner[] = [];
  let tenants: Tenant[] = [];
  let errorMessage = "";
  try {
    [settings, members, links, owners, tenants] = await Promise.all([
      getOrganizationSettings(),
      listOrganizationMembers(),
      listPortalLinks(),
      listOwners(),
      listTenants(),
    ]);
  } catch (error) {
    errorMessage = error instanceof AdminConfigurationError ? error.message : "Organization administration data could not be loaded. Confirm the Go API and latest migrations are running.";
  }

  const ownerUsers = members.filter((item) => item.role === "owner");
  const tenantUsers = members.filter((item) => item.role === "tenant");

  return (
    <AppShell section="Settings">
      <header className="workspaceHeader">
        <div><p className="eyebrow">CONTROL PLANE</p><h1 className="pageTitle">Organization settings</h1><p className="pageIntro">Manage operating defaults, team access and the explicit links that power owner and tenant portals.</p></div>
        <div className="metricStrip"><div><strong>{members.length}</strong><span>Members</span></div><div><strong>{links.length}</strong><span>Portal links</span></div><div><strong>{members.filter((item) => item.role === "admin").length}</strong><span>Admins</span></div></div>
      </header>

      {errorMessage ? <div className="noticePanel"><strong>Administration unavailable</strong><p>{errorMessage}</p></div> : null}

      <div className="leasingGrid">
        <section className="panel">
          <div className="panelHeader"><div><p className="panelKicker">ORGANIZATION</p><h2>Operating defaults</h2></div></div>
          <form className="stackedForm" action={updateOrganizationAction}>
            <label>Organization name<input name="name" required defaultValue={settings?.name ?? ""} /></label>
            <div className="formPair"><label>Timezone<input name="timezone" required defaultValue={settings?.timezone ?? "Asia/Colombo"} /></label><label>Default currency<input name="defaultCurrency" required maxLength={3} defaultValue={settings?.defaultCurrency ?? "LKR"} /></label></div>
            <div className="formPair"><label>Country code<input name="countryCode" required maxLength={2} defaultValue={settings?.countryCode ?? "LK"} /></label><label>Billing email<input name="billingEmail" type="email" defaultValue={settings?.billingEmail ?? ""} /></label></div>
            <button className="primaryButton" disabled={!settings}>Save organization settings</button>
          </form>
        </section>

        <section className="panel">
          <div className="panelHeader"><div><p className="panelKicker">TEAM</p><h2>Add or invite member</h2></div></div>
          <form className="stackedForm" action={createMemberAction}>
            <label>Email<input name="email" type="email" required placeholder="person@company.com" /></label>
            <label>Display name<input name="displayName" required placeholder="Property manager" /></label>
            <label>Role<select name="role" defaultValue="viewer">{roles.map((role) => <option key={role} value={role}>{role}</option>)}</select></label>
            <button className="primaryButton">Add member</button>
          </form>
        </section>
      </div>

      <section className="panel leasingRegister">
        <div className="panelHeader"><div><p className="panelKicker">ACCESS</p><h2>Organization members</h2></div><span className="countBadge">{members.length}</span></div>
        {members.length ? <div className="dataTableWrap"><table className="dataTable"><thead><tr><th>Member</th><th>Status</th><th>Role</th><th>Change role</th><th>Remove</th></tr></thead><tbody>
          {members.map((member) => <tr key={member.userId}><td><strong>{member.displayName}</strong><span className="recordMeta">{member.email}</span></td><td><span className={`statusPill status-${member.userStatus}`}>{member.userStatus}</span></td><td className="capitalize">{member.role}</td><td><form action={updateMemberRoleAction} className="inlineForm"><input type="hidden" name="userId" value={member.userId} /><select name="role" defaultValue={member.role}>{roles.map((role) => <option key={role} value={role}>{role}</option>)}</select><button className="tableButton">Save</button></form></td><td><form action={removeMemberAction}><input type="hidden" name="userId" value={member.userId} /><button className="tableButton tableButtonDanger">Remove</button></form></td></tr>)}
        </tbody></table></div> : <div className="emptyState"><h3>No members</h3><p>Add an administrator before using the organization.</p></div>}
      </section>

      <div className="leasingGrid">
        <section className="panel">
          <div className="panelHeader"><div><p className="panelKicker">OWNER PORTAL</p><h2>Link login to owner</h2></div></div>
          <form className="stackedForm" action={linkOwnerPortalAction}>
            <label>Owner-role member<select name="userId" required defaultValue=""><option value="" disabled>Select user</option>{ownerUsers.map((item) => <option value={item.userId} key={item.userId}>{item.displayName} · {item.email}</option>)}</select></label>
            <label>Owner record<select name="ownerId" required defaultValue=""><option value="" disabled>Select owner</option>{owners.filter((item) => item.status === "active").map((item) => <option value={item.id} key={item.id}>{item.legalName}</option>)}</select></label>
            <button className="primaryButton" disabled={!ownerUsers.length || !owners.length}>Link owner portal</button>
          </form>
        </section>
        <section className="panel">
          <div className="panelHeader"><div><p className="panelKicker">TENANT PORTAL</p><h2>Link login to tenant</h2></div></div>
          <form className="stackedForm" action={linkTenantPortalAction}>
            <label>Tenant-role member<select name="userId" required defaultValue=""><option value="" disabled>Select user</option>{tenantUsers.map((item) => <option value={item.userId} key={item.userId}>{item.displayName} · {item.email}</option>)}</select></label>
            <label>Tenant record<select name="tenantId" required defaultValue=""><option value="" disabled>Select tenant</option>{tenants.filter((item) => item.status !== "blocked").map((item) => <option value={item.id} key={item.id}>{item.legalName}</option>)}</select></label>
            <button className="primaryButton" disabled={!tenantUsers.length || !tenants.length}>Link tenant portal</button>
          </form>
        </section>
      </div>

      <section className="panel leasingRegister">
        <div className="panelHeader"><div><p className="panelKicker">PORTAL ACCESS</p><h2>Active resource links</h2></div><span className="countBadge">{links.length}</span></div>
        {links.length ? <div className="dataTableWrap"><table className="dataTable"><thead><tr><th>Portal</th><th>User</th><th>Linked record</th><th>Action</th></tr></thead><tbody>{links.map((link) => <tr key={`${link.kind}:${link.userId}:${link.resourceId}`}><td className="capitalize">{link.kind}</td><td><strong>{link.displayName}</strong><span className="recordMeta">{link.email}</span></td><td>{link.resourceName}</td><td><form action={link.kind === "owner" ? unlinkOwnerPortalAction : unlinkTenantPortalAction}><input type="hidden" name="userId" value={link.userId} /><input type="hidden" name="resourceId" value={link.resourceId} /><button className="tableButton tableButtonDanger">Unlink</button></form></td></tr>)}</tbody></table></div> : <div className="emptyState"><h3>No portal links</h3><p>Owner and tenant users need explicit links before their portal can return business data.</p></div>}
      </section>
    </AppShell>
  );
}
