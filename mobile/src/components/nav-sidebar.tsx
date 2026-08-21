import { Ionicons } from '@expo/vector-icons';
import type { DrawerContentComponentProps } from '@react-navigation/drawer';
import { DrawerContentScrollView } from '@react-navigation/drawer';
import { Pressable, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { Avatar } from '@/components/ui/avatar';
import { LogoutButton } from '@/components/logout-button';
import { useAuthStore } from '@/stores/auth-store';

type IconName = keyof typeof Ionicons.glyphMap;

// Keyed by expo-router route filename (matches Drawer.Screen `name`), not by
// label — labels come from each Drawer.Screen's `options.title` so this map
// only needs to stay in sync when a route file is added or renamed.
const ROUTE_ICONS: Record<string, IconName> = {
  dashboard: 'home-outline',
  students: 'people-outline',
  activities: 'document-text-outline',
  attendance: 'checkmark-done-circle-outline',
  reports: 'bar-chart-outline',
  supervision: 'people-circle-outline',
  evaluations: 'star-outline',
  competencies: 'ribbon-outline',
  resources: 'book-outline',
  notifications: 'notifications-outline',
};

export function NavSidebar(props: DrawerContentComponentProps) {
  const { state, descriptors, navigation } = props;
  const user = useAuthStore((s) => s.user);

  return (
    <SafeAreaView className="flex-1 bg-white" edges={['top', 'bottom']}>
      <DrawerContentScrollView {...props} contentContainerStyle={{ paddingTop: 0 }}>
        <View className="mb-2 gap-1 border-b border-slate-100 px-6 pb-6">
          <View className="mb-3 h-11 w-11 items-center justify-center rounded-2xl bg-brand-600">
            <Ionicons name="leaf-outline" size={21} color="#ffffff" />
          </View>
          <Text className="text-lg font-bold text-slate-900">FieldSync</Text>
          <Text className="text-xs text-slate-400">Field Practicum Management</Text>
        </View>

        <View className="px-3 pt-3">
          {state.routes.map((route, index) => {
            const { options } = descriptors[route.key]!;
            const label = (options.title ?? route.name) as string;
            const focused = state.index === index;
            const icon = ROUTE_ICONS[route.name] ?? 'ellipse-outline';

            return (
              <Pressable
                key={route.key}
                onPress={() => navigation.navigate(route.name)}
                className={`mb-0.5 flex-row items-center gap-3 rounded-xl px-4 py-3 active:bg-slate-50 ${
                  focused ? 'bg-brand-50' : ''
                }`}
              >
                <Ionicons name={icon} size={19} color={focused ? '#4338ca' : '#64748b'} />
                <Text className={`text-[15px] font-medium ${focused ? 'text-brand-700' : 'text-slate-600'}`}>
                  {label}
                </Text>
              </Pressable>
            );
          })}
        </View>
      </DrawerContentScrollView>

      <View className="border-t border-slate-100 px-4 pb-2 pt-4">
        <View className="mb-3 flex-row items-center gap-3 px-2">
          <Avatar name={user?.fullName ?? '?'} size="sm" />
          <View className="flex-1">
            <Text numberOfLines={1} className="text-sm font-semibold text-slate-800">
              {user?.fullName ?? 'Account'}
            </Text>
            <Text className="text-xs capitalize text-slate-400">{user?.role.replace(/_/g, ' ')}</Text>
          </View>
        </View>
        <LogoutButton fullWidth />
      </View>
    </SafeAreaView>
  );
}
