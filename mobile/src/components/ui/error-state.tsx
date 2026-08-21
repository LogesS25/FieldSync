import { Ionicons } from '@expo/vector-icons';
import { Pressable, Text, View } from 'react-native';

interface ErrorStateProps {
  title?: string;
  description?: string;
  onRetry?: () => void;
}

// Shared "the request failed" treatment. Distinct from EmptyState (which
// means "the request succeeded and there's genuinely nothing here") — most
// screens previously only handled the empty case and silently showed an
// empty-state message even on a network error, which is misleading.
export function ErrorState({
  title = 'Something went wrong',
  description = 'Could not load this. Check your connection and try again.',
  onRetry,
}: ErrorStateProps) {
  return (
    <View className="items-center rounded-2xl border border-rose-100 bg-rose-50/60 px-6 py-10">
      <View className="mb-3 h-11 w-11 items-center justify-center rounded-full bg-rose-100">
        <Ionicons name="alert-circle-outline" size={22} color="#e11d48" />
      </View>
      <Text className="text-base font-semibold text-slate-800">{title}</Text>
      <Text className="mt-1 text-center text-sm text-slate-500">{description}</Text>
      {onRetry ? (
        <Pressable
          onPress={onRetry}
          className="mt-4 flex-row items-center gap-1.5 rounded-full border border-rose-200 bg-white px-4 py-2 active:opacity-70"
        >
          <Ionicons name="refresh-outline" size={15} color="#e11d48" />
          <Text className="text-sm font-semibold text-rose-600">Try again</Text>
        </Pressable>
      ) : null}
    </View>
  );
}
