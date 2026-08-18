import { Tabs } from 'expo-router';

// Screens correspond to the Student module capabilities in the requirements
// (§4.1, §5.2, §12): practicum overview, field work logging, attendance,
// weekly reports, supervision info, competency progress, notifications,
// and the standardized resource library.
export default function StudentLayout() {
  return (
    <Tabs>
      <Tabs.Screen name="dashboard" options={{ title: 'Dashboard' }} />
      <Tabs.Screen name="activities" options={{ title: 'Activities' }} />
      <Tabs.Screen name="attendance" options={{ title: 'Attendance' }} />
      <Tabs.Screen name="reports" options={{ title: 'Reports' }} />
      <Tabs.Screen name="supervision" options={{ title: 'Supervision' }} />
      <Tabs.Screen name="competencies" options={{ title: 'Competencies' }} />
      <Tabs.Screen name="resources" options={{ title: 'Resources' }} />
      <Tabs.Screen name="notifications" options={{ title: 'Notifications' }} />
    </Tabs>
  );
}
