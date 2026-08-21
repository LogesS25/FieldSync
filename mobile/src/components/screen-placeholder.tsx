import { Ionicons } from '@expo/vector-icons';
import type { PropsWithChildren } from 'react';
import { Text, View } from 'react-native';

import { ScreenContainer } from '@/components/ui/screen-container';

interface ScreenPlaceholderProps extends PropsWithChildren {
  title: string;
  note?: string;
  icon?: keyof typeof Ionicons.glyphMap;
}

// Shared placeholder for screens whose real implementation lands in a later
// phase (see docs/ARCHITECTURE.md §10). Styled as an intentional "coming
// soon" state — not a bare stub — so a not-yet-built screen still feels like
// part of a finished app rather than a broken one.
export function ScreenPlaceholder({ title, note, icon = 'construct-outline', children }: ScreenPlaceholderProps) {
  return (
    <ScreenContainer>
      <View className="flex-1 items-center justify-center px-6">
        <View className="mb-4 h-16 w-16 items-center justify-center rounded-3xl bg-brand-50">
          <Ionicons name={icon} size={28} color="#4f46e5" />
        </View>
        <Text className="text-center text-xl font-bold text-slate-900">{title}</Text>
        {note ? <Text className="mt-2 text-center text-sm leading-5 text-slate-500">{note}</Text> : null}
        {children}
      </View>
    </ScreenContainer>
  );
}
