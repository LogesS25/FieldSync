import { Text, View } from 'react-native';

interface EmptyStateProps {
  title: string;
  description?: string;
}

export function EmptyState({ title, description }: EmptyStateProps) {
  return (
    <View className="items-center rounded-2xl border border-dashed border-slate-200 bg-white/60 px-6 py-10">
      <Text className="text-base font-semibold text-slate-700">{title}</Text>
      {description ? (
        <Text className="mt-1 text-center text-sm text-slate-500">{description}</Text>
      ) : null}
    </View>
  );
}
