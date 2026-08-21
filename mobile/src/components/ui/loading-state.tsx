import { ActivityIndicator, Text, View } from 'react-native';

interface LoadingStateProps {
  message?: string;
  compact?: boolean;
}

// Shared loading treatment so every screen's "waiting on the network" moment
// looks the same instead of each screen improvising its own ActivityIndicator
// placement.
export function LoadingState({ message, compact }: LoadingStateProps) {
  return (
    <View className={compact ? 'items-center py-6' : 'items-center py-16'}>
      <ActivityIndicator size="small" color="#4f46e5" />
      {message ? <Text className="mt-3 text-sm text-slate-400">{message}</Text> : null}
    </View>
  );
}
