import type { DrawerNavigationOptions } from '@react-navigation/drawer';
import { useWindowDimensions } from 'react-native';

const WIDE_BREAKPOINT = 768;

// Shared Drawer chrome for both role layouts: a permanently visible sidebar
// on tablet/web-width screens, a slide-out overlay on phone width, and a
// consistent branded top bar (Drawer's own header) everywhere.
export function useDrawerScreenOptions(): DrawerNavigationOptions {
  const { width } = useWindowDimensions();
  const isWide = width >= WIDE_BREAKPOINT;

  return {
    headerShown: true,
    headerStyle: { backgroundColor: '#ffffff' },
    headerShadowVisible: false,
    headerTintColor: '#1e293b',
    headerTitleStyle: { fontWeight: '700', fontSize: 17 },
    drawerType: isWide ? 'permanent' : 'front',
    drawerStyle: { width: isWide ? 288 : 296 },
    overlayColor: 'rgba(15, 23, 42, 0.4)',
    sceneStyle: { backgroundColor: '#f8fafc' },
  };
}
