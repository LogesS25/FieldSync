import { ScreenPlaceholder } from '@/components/screen-placeholder';
import { LogoutButton } from '@/components/logout-button';
import { useAuthStore } from '@/stores/auth-store';

export default function StudentDashboard() {
  const user = useAuthStore((state) => state.user);

  return (
    <ScreenPlaceholder
      title={`Welcome, ${user?.fullName ?? 'Student'}`}
      note="Full dashboard built in Phase 4 (Student Fieldwork)."
    >
      <LogoutButton />
    </ScreenPlaceholder>
  );
}
