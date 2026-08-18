import { Text, View } from 'react-native';

interface ScreenPlaceholderProps {
  title: string;
  note?: string;
}

// Shared placeholder for screens whose real implementation lands in a later
// phase (see docs/ARCHITECTURE.md §10). Keeps route scaffolding for Phase 1
// without stubbing out business logic ahead of time.
export function ScreenPlaceholder({ title, note }: ScreenPlaceholderProps) {
  return (
    <View className="flex-1 items-center justify-center bg-white px-6">
      <Text className="text-xl font-semibold text-slate-900">{title}</Text>
      {note ? <Text className="mt-2 text-center text-slate-500">{note}</Text> : null}
    </View>
  );
}
