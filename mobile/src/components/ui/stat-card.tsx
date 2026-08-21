import { Ionicons } from '@expo/vector-icons';
import { Pressable, Text, View } from 'react-native';

interface StatCardProps {
  label: string;
  value: number | string;
  icon: keyof typeof Ionicons.glyphMap;
  onPress?: () => void;
}

export function StatCard({ label, value, icon, onPress }: StatCardProps) {
  return (
    <Pressable
      onPress={onPress}
      disabled={!onPress}
      className="min-w-[150px] flex-1 rounded-2xl border border-slate-100 bg-white p-4 shadow-sm shadow-slate-200 active:opacity-80"
    >
      <View className="mb-3 h-9 w-9 items-center justify-center rounded-xl bg-brand-50">
        <Ionicons name={icon} size={17} color="#4f46e5" />
      </View>
      <Text className="text-2xl font-bold text-slate-900">{value}</Text>
      <Text className="mt-0.5 text-xs font-medium text-slate-500">{label}</Text>
    </Pressable>
  );
}
