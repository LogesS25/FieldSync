import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { Link, router } from 'expo-router';
import { Controller, useForm } from 'react-hook-form';
import { Pressable, Text, TextInput, View } from 'react-native';

import { loginSchema, type LoginFormValues } from '@/features/auth/schemas';
import { ApiError } from '@/lib/api-client';
import * as authService from '@/services/auth';
import { useAuthStore } from '@/stores/auth-store';

export default function LoginScreen() {
  const setSession = useAuthStore((state) => state.setSession);

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: '', password: '' },
  });

  const mutation = useMutation({
    mutationFn: (values: LoginFormValues) => authService.login(values.email, values.password),
    onSuccess: (data) => {
      setSession(data.user, data.accessToken, data.refreshToken);
      // Updating the store alone doesn't move the router off the login
      // screen — route to "/" and let index.tsx's role-based redirect send
      // the user to the right dashboard from one place.
      router.replace('/');
    },
  });

  const onSubmit = handleSubmit((values) => mutation.mutate(values));

  return (
    <View className="flex-1 justify-center bg-white px-6">
      <Text className="mb-1 text-2xl font-semibold text-slate-900">FieldSync</Text>
      <Text className="mb-8 text-slate-500">Sign in to continue your field practicum.</Text>

      <Text className="mb-1 text-sm font-medium text-slate-700">Email</Text>
      <Controller
        control={control}
        name="email"
        render={({ field: { onChange, onBlur, value } }) => (
          <TextInput
            className="mb-1 rounded-lg border border-slate-300 px-4 py-3 text-slate-900"
            autoCapitalize="none"
            autoComplete="email"
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
            autoComplete="password"
            secureTextEntry
            placeholder="••••••••"
            onBlur={onBlur}
            onChangeText={onChange}
            value={value}
          />
        )}
      />
      {errors.password ? <Text className="mb-2 text-sm text-red-600">{errors.password.message}</Text> : null}

      {mutation.isError ? (
        <Text className="mb-2 text-sm text-red-600">
          {mutation.error instanceof ApiError && mutation.error.status === 401
            ? 'Incorrect email or password.'
            : 'Something went wrong. Please try again.'}
        </Text>
      ) : null}

      <Pressable
        className="mt-4 items-center rounded-lg bg-brand-600 py-3 disabled:opacity-50"
        disabled={mutation.isPending}
        onPress={onSubmit}
      >
        <Text className="font-semibold text-white">{mutation.isPending ? 'Signing in…' : 'Sign In'}</Text>
      </Pressable>

      <Link href="/(auth)/register" className="mt-6 text-center text-brand-600">
        Don&apos;t have an account? Register
      </Link>
    </View>
  );
}
