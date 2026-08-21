import { Text, TextInput, type TextInputProps, View } from 'react-native';

interface LabeledInputProps extends TextInputProps {
  label: string;
  error?: string;
}

export function LabeledInput({ label, error, ...inputProps }: LabeledInputProps) {
  return (
    <View className="mb-4">
      <Text className="mb-2 text-sm font-medium text-slate-700">{label}</Text>
      <TextInput
        className={`rounded-xl border bg-white px-4 py-3.5 text-slate-900 ${
          error ? 'border-rose-300' : 'border-slate-200'
        }`}
        placeholderTextColor="#94a3b8"
        {...inputProps}
      />
      {error ? <Text className="mt-1.5 text-sm text-red-600">{error}</Text> : null}
    </View>
  );
}
