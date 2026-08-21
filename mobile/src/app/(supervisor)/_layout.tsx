import { Drawer } from 'expo-router/drawer';

import { NavSidebar } from '@/components/nav-sidebar';
import { RequireRole } from '@/components/require-role';
import { useDrawerScreenOptions } from '@/lib/use-drawer-screen-options';

// Shared by Faculty and Agency Supervisors (requirements §4.2/§4.3, §12) —
// both roles view assigned students, verify/review records, record
// supervision, evaluate, and give feedback. Role-specific differences
// (e.g. grievances are agency-only) are gated inside each screen, not by
// separate route groups, since screens/permissions overlap far more than
// they diverge.
export default function SupervisorLayout() {
  const screenOptions = useDrawerScreenOptions();

  return (
    <RequireRole allowedRoles={['faculty_supervisor', 'agency_supervisor']}>
      <Drawer screenOptions={screenOptions} drawerContent={(props) => <NavSidebar {...props} />}>
        <Drawer.Screen name="dashboard" options={{ title: 'Dashboard' }} />
        <Drawer.Screen name="students" options={{ title: 'Students' }} />
        <Drawer.Screen name="activities" options={{ title: 'Daily Reports' }} />
        <Drawer.Screen name="attendance" options={{ title: 'Attendance' }} />
        <Drawer.Screen name="supervision" options={{ title: 'Team Requests' }} />
        <Drawer.Screen name="evaluations" options={{ title: 'Evaluations' }} />
        <Drawer.Screen name="resources" options={{ title: 'Resources' }} />
        <Drawer.Screen name="notifications" options={{ title: 'Notifications' }} />
      </Drawer>
    </RequireRole>
  );
}
