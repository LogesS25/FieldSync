import { Ionicons } from '@expo/vector-icons';
import { Text, View } from 'react-native';

interface EmptyStateProps {
  title: string;
  description?: string;
  icon?: keyof typeof Ionicons.glyphMap;
}

export function EmptyState({ title, description, icon = 'file-tray-outline' }: EmptyStateProps) {
  return (
    <View className="items-center rounded-2xl border border-dashed border-slate-200 bg-white/60 px-6 py-10">
      <View className="mb-3 h-11 w-11 items-center justify-center rounded-full bg-slate-100">
        <Ionicons name={icon} size={20} color="#94a3b8" />
      </View>
      <Text className="text-base font-semibold text-slate-700">{title}</Text>
      {description ? (
        <Text className="mt-1 text-center text-sm text-slate-500">{description}</Text>
      ) : null}
    </View>
  );
}
