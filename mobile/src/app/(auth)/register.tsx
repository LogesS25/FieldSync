import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { Link, router } from 'expo-router';
import { Controller, useForm } from 'react-hook-form';
import { Pressable, ScrollView, Text, TextInput, View } from 'react-native';

import { registerSchema, type RegisterFormValues } from '@/features/auth/schemas';
import { ApiError } from '@/lib/api-client';
import * as authService from '@/services/auth';
import { useAuthStore } from '@/stores/auth-store';
import type { UserRole } from '@/types/auth';

const ROLE_OPTIONS: { value: UserRole; label: string }[] = [
  { value: 'student', label: 'Student' },
  { value: 'faculty_supervisor', label: 'Faculty Supervisor' },
  { value: 'agency_supervisor', label: 'Agency Supervisor' },
];

export default function RegisterScreen() {
  const setSession = useAuthStore((state) => state.setSession);

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: { fullName: '', email: '', password: '', role: 'student' },
  });

  const mutation = useMutation({
    mutationFn: authService.register,
    onSuccess: (data) => {
      setSession(data.user, data.accessToken, data.refreshToken);
      router.replace('/');
    },
  });

  const onSubmit = handleSubmit((values) => mutation.mutate(values));

  return (
    <ScrollView className="flex-1 bg-white" contentContainerClassName="justify-center px-6 py-12">
      <Text className="mb-1 text-2xl font-semibold text-slate-900">Create your account</Text>
      <Text className="mb-8 text-slate-500">Register as a Student or Supervisor.</Text>

      <Text className="mb-1 text-sm font-medium text-slate-700">Full name</Text>
      <Controller
        control={control}
        name="fullName"
        render={({ field: { onChange, onBlur, value } }) => (
          <TextInput
            className="mb-1 rounded-lg border border-slate-300 px-4 py-3 text-slate-900"
            placeholder="Jane Doe"
            onBlur={onBlur}
            onChangeText={onChange}
            value={value}
          />
        )}
      />
      {errors.fullName ? <Text className="mb-2 text-sm text-red-600">{errors.fullName.message}</Text> : null}

      <Text className="mb-1 mt-3 text-sm font-medium text-slate-700">Email</Text>
      <Controller
        control={control}
        name="email"
        render={({ field: { onChange, onBlur, value } }) => (
          <TextInput
            className="mb-1 rounded-lg border border-slate-300 px-4 py-3 text-slate-900"
            autoCapitalize="none"
            keyboardType="email-address"
            placeholder="you@university.edu"
            onBlur={onBlur}
            onChangeText={onChange}
            value={value}
          />
        )}
      />
      {errors.email ? <Text className="mb-2 text-sm text-red-600">{errors.email.message}</Text> : null}

      <Text className="mb-1 mt-3 text-sm font-medium text-slate-700">Password</Text>
      <Controller
        control={control}
        name="password"
        render={({ field: { onChange, onBlur, value } }) => (
          <TextInput
            className="mb-1 rounded-lg border border-slate-300 px-4 py-3 text-slate-900"
            autoCapitalize="none"
            secureTextEntry
            placeholder="At least 8 characters"
            onBlur={onBlur}
            onChangeText={onChange}
            value={value}
          />
        )}
      />
      {errors.password ? <Text className="mb-2 text-sm text-red-600">{errors.password.message}</Text> : null}

      <Text className="mb-2 mt-3 text-sm font-medium text-slate-700">I am a</Text>
      <Controller
        control={control}
        name="role"
        render={({ field: { onChange, value } }) => (
          <View className="mb-2 flex-row flex-wrap gap-2">
            {ROLE_OPTIONS.map((option) => (
              <Pressable
                key={option.value}
                onPress={() => onChange(option.value)}
                className={`rounded-full border px-4 py-2 ${
                  value === option.value ? 'border-brand-600 bg-brand-600' : 'border-slate-300 bg-white'
                }`}
              >
                <Text className={value === option.value ? 'text-white' : 'text-slate-700'}>{option.label}</Text>
              </Pressable>
            ))}
          </View>
        )}
      />

      {mutation.isError ? (
        <Text className="mb-2 text-sm text-red-600">
          {mutation.error instanceof ApiError && mutation.error.status === 409
            ? 'That email is already registered.'
            : 'Something went wrong. Please try again.'}
        </Text>
      ) : null}

      <Pressable
        className="mt-4 items-center rounded-lg bg-brand-600 py-3 disabled:opacity-50"
        disabled={mutation.isPending}
        onPress={onSubmit}
      >
        <Text className="font-semibold text-white">{mutation.isPending ? 'Creating account…' : 'Register'}</Text>
      </Pressable>

      <Link href="/(auth)/login" className="mt-6 text-center text-brand-600">
        Already have an account? Sign in
      </Link>
    </ScrollView>
  );
}
