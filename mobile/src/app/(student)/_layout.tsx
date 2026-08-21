import { Tabs } from 'expo-router';

import { RequireRole } from '@/components/require-role';

// Screens correspond to the Student module capabilities in the requirements
// (§4.1, §5.2, §12): practicum overview, field work logging, attendance,
// weekly reports, supervision info, competency progress, notifications,
// and the standardized resource library.
export default function StudentLayout() {
  return (
    <RequireRole allowedRoles={['student']}>
      <Tabs>
        <Tabs.Screen name="dashboard" options={{ title: 'Dashboard' }} />
        <Tabs.Screen name="activities" options={{ title: 'Daily Reports' }} />
        <Tabs.Screen name="attendance" options={{ title: 'Attendance' }} />
        <Tabs.Screen name="reports" options={{ title: 'Reports' }} />
        <Tabs.Screen name="supervision" options={{ title: 'Supervision' }} />
        <Tabs.Screen name="competencies" options={{ title: 'Competencies' }} />
        <Tabs.Screen name="resources" options={{ title: 'Resources' }} />
        <Tabs.Screen name="notifications" options={{ title: 'Notifications' }} />
      </Tabs>
    </RequireRole>
  );
}
