import { Pressable, Text, View } from 'react-native';

interface PillOption<T extends string> {
  value: T;
  label: string;
}

interface PillSelectProps<T extends string> {
  options: PillOption<T>[];
  value: T | '';
  onChange: (value: T) => void;
}

// Shared "pick one of a small set" control (role, agency, session, decision
// pickers, etc.) so every screen with this pattern looks and behaves the
// same instead of each reimplementing the same Pressable/View markup.
export function PillSelect<T extends string>({ options, value, onChange }: PillSelectProps<T>) {
  return (
    <View className="flex-row flex-wrap gap-2">
      {options.map((option) => {
        const selected = value === option.value;
        return (
          <Pressable
            key={option.value}
            onPress={() => onChange(option.value)}
            className={`rounded-full border px-4 py-2.5 active:opacity-80 ${
              selected ? 'border-brand-600 bg-brand-600' : 'border-slate-200 bg-white'
            }`}
          >
            <Text className={`text-sm font-medium ${selected ? 'text-white' : 'text-slate-600'}`}>{option.label}</Text>
          </Pressable>
        );
      })}
    </View>
  );
}
