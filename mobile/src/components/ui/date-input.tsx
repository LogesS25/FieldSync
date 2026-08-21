import DateTimePicker from '@react-native-community/datetimepicker';
import { Ionicons } from '@expo/vector-icons';
import { useState } from 'react';
import { Modal, Platform, Pressable, Text, View } from 'react-native';

interface DateInputProps {
  label: string;
  value: string; // YYYY-MM-DD
  onChange: (value: string) => void;
  error?: string;
}

function toDate(value: string): Date {
  return value ? new Date(`${value}T00:00:00`) : new Date();
}

function formatISO(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

function formatDisplay(value: string): string {
  return toDate(value).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
}

// Android's picker is a native dialog (open → pick → auto-closes), so it
// needs no wrapper. iOS has no equivalent modal dialog for date-only
// pickers, so we wrap its spinner in our own bottom-sheet with
// Cancel/Done — otherwise there's no way to dismiss it.
export function DateInput({ label, value, onChange, error }: DateInputProps) {
  const [show, setShow] = useState(false);
  const [draft, setDraft] = useState(() => toDate(value));

  const openPicker = () => {
    setDraft(toDate(value));
    setShow(true);
  };

  const trigger = (
    <Pressable
      onPress={openPicker}
      className={`flex-row items-center justify-between rounded-xl border bg-white px-4 py-3.5 ${
        error ? 'border-rose-300' : 'border-slate-200'
      }`}
    >
      <Text className={value ? 'text-slate-900' : 'text-slate-400'}>{value ? formatDisplay(value) : 'Select date'}</Text>
      <Ionicons name="calendar-outline" size={17} color="#94a3b8" />
    </Pressable>
  );

  if (Platform.OS === 'android') {
    return (
      <View className="mb-4">
        <Text className="mb-2 text-sm font-medium text-slate-700">{label}</Text>
        {trigger}
        {error ? <Text className="mt-1.5 text-sm text-red-600">{error}</Text> : null}
        {show ? (
          <DateTimePicker
            value={draft}
            mode="date"
            display="default"
            onChange={(event, selectedDate) => {
              setShow(false);
              if (event.type === 'set' && selectedDate) {
                onChange(formatISO(selectedDate));
              }
            }}
          />
        ) : null}
      </View>
    );
  }

  return (
    <View className="mb-4">
      <Text className="mb-2 text-sm font-medium text-slate-700">{label}</Text>
      {trigger}
      {error ? <Text className="mt-1.5 text-sm text-red-600">{error}</Text> : null}

      <Modal visible={show} transparent animationType="fade" onRequestClose={() => setShow(false)}>
        <View className="flex-1 items-center justify-end bg-black/40">
          <View className="w-full rounded-t-2xl bg-white p-4 pb-8">
            <DateTimePicker
              value={draft}
              mode="date"
              display="spinner"
              onChange={(_, selectedDate) => {
                if (selectedDate) setDraft(selectedDate);
              }}
            />
            <View className="mt-2 flex-row justify-end gap-3">
              <Pressable onPress={() => setShow(false)} className="px-4 py-2">
                <Text className="text-slate-500">Cancel</Text>
              </Pressable>
              <Pressable
                onPress={() => {
                  onChange(formatISO(draft));
                  setShow(false);
                }}
                className="rounded-xl bg-brand-600 px-5 py-2.5 active:bg-brand-700"
              >
                <Text className="font-semibold text-white">Done</Text>
              </Pressable>
            </View>
          </View>
        </View>
      </Modal>
    </View>
  );
}
