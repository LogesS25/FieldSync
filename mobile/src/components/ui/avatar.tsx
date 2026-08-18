import { Text, View } from 'react-native';

interface AvatarProps {
  name: string;
  size?: 'sm' | 'md' | 'lg';
}

const SIZE_STYLES: Record<NonNullable<AvatarProps['size']>, { container: string; text: string }> = {
  sm: { container: 'h-8 w-8', text: 'text-xs' },
  md: { container: 'h-11 w-11', text: 'text-sm' },
  lg: { container: 'h-14 w-14', text: 'text-base' },
};

function initialsFor(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return '?';
  if (parts.length === 1) return parts[0]!.slice(0, 2).toUpperCase();
  return (parts[0]![0]! + parts[parts.length - 1]![0]!).toUpperCase();
}

export function Avatar({ name, size = 'md' }: AvatarProps) {
  const styles = SIZE_STYLES[size];
  return (
    <View className={`items-center justify-center rounded-full bg-brand-100 ${styles.container}`}>
      <Text className={`font-semibold text-brand-700 ${styles.text}`}>{initialsFor(name)}</Text>
    </View>
  );
}
