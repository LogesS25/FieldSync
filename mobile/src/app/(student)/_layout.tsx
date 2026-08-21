import { Drawer } from 'expo-router/drawer';

import { NavSidebar } from '@/components/nav-sidebar';
import { RequireRole } from '@/components/require-role';
import { useDrawerScreenOptions } from '@/lib/use-drawer-screen-options';

// Screens correspond to the Student module capabilities in the requirements
// (§4.1, §5.2, §12): practicum overview, daily reports, attendance,
// consolidated reports, supervision/team, competency progress,
// notifications, and the standardized resource library.
export default function StudentLayout() {
  const screenOptions = useDrawerScreenOptions();

  return (
    <RequireRole allowedRoles={['student']}>
      <Drawer screenOptions={screenOptions} drawerContent={(props) => <NavSidebar {...props} />}>
        <Drawer.Screen name="dashboard" options={{ title: 'Dashboard' }} />
        <Drawer.Screen name="activities" options={{ title: 'Daily Reports' }} />
        <Drawer.Screen name="attendance" options={{ title: 'Attendance' }} />
        <Drawer.Screen name="reports" options={{ title: 'Reports' }} />
        <Drawer.Screen name="supervision" options={{ title: 'Team' }} />
        <Drawer.Screen name="competencies" options={{ title: 'Competencies' }} />
        <Drawer.Screen name="resources" options={{ title: 'Resources' }} />
        <Drawer.Screen name="notifications" options={{ title: 'Notifications' }} />
      </Drawer>
    </RequireRole>
  );
}
