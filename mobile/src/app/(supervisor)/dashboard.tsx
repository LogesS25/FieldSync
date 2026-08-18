import { ScreenPlaceholder } from '@/components/screen-placeholder';
import { LogoutButton } from '@/components/logout-button';
import { useAuthStore } from '@/stores/auth-store';

export default function SupervisorDashboard() {
  const user = useAuthStore((state) => state.user);

  return (
    <ScreenPlaceholder
      title={`Welcome, ${user?.fullName ?? 'Supervisor'}`}
      note="Full dashboard built in Phase 5 (Supervisor Workflows)."
    >
      <LogoutButton />
    </ScreenPlaceholder>
  );
}
