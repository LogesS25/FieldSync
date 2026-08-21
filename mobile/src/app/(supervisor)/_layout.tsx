import { Tabs } from 'expo-router';

import { RequireRole } from '@/components/require-role';

// Shared by Faculty and Agency Supervisors (requirements §4.2/§4.3, §12) —
// both roles view assigned students, verify/review records, record
// supervision, evaluate, and give feedback. Role-specific differences
// (e.g. grievances are agency-only) are gated inside each screen, not by
// separate route groups, since screens/permissions overlap far more than
// they diverge.
export default function SupervisorLayout() {
  return (
    <RequireRole allowedRoles={['faculty_supervisor', 'agency_supervisor']}>
      <Tabs>
        <Tabs.Screen name="dashboard" options={{ title: 'Dashboard' }} />
        <Tabs.Screen name="students" options={{ title: 'Students' }} />
        <Tabs.Screen name="activities" options={{ title: 'Daily Reports' }} />
        <Tabs.Screen name="attendance" options={{ title: 'Attendance' }} />
        <Tabs.Screen name="supervision" options={{ title: 'Supervision' }} />
        <Tabs.Screen name="evaluations" options={{ title: 'Evaluations' }} />
        <Tabs.Screen name="resources" options={{ title: 'Resources' }} />
        <Tabs.Screen name="notifications" options={{ title: 'Notifications' }} />
      </Tabs>
    </RequireRole>
  );
}
