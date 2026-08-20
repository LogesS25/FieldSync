import { Pressable, Text } from 'react-native';

interface PrimaryButtonProps {
  label: string;
  onPress: () => void;
  disabled?: boolean;
}

export function PrimaryButton({ label, onPress, disabled }: PrimaryButtonProps) {
  return (
    <Pressable
      className="items-center rounded-lg bg-brand-600 py-3 disabled:opacity-50"
      disabled={disabled}
      onPress={onPress}
    >
      <Text className="font-semibold text-white">{label}</Text>
    </Pressable>
  );
}
