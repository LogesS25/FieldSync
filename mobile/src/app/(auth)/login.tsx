import { zodResolver } from '@hookform/resolvers/zod';
import { Ionicons } from '@expo/vector-icons';
import { useMutation } from '@tanstack/react-query';
import { Link, router } from 'expo-router';
import { Controller, useForm } from 'react-hook-form';
import { KeyboardAvoidingView, Platform, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { LabeledInput } from '@/components/ui/labeled-input';
import { PrimaryButton } from '@/components/ui/primary-button';
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
    <SafeAreaView className="flex-1 bg-white">
      <KeyboardAvoidingView className="flex-1" behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
        <View className="flex-1 justify-center px-7">
          <View className="mb-10 items-center">
            <View className="mb-4 h-14 w-14 items-center justify-center rounded-2xl bg-brand-600">
              <Ionicons name="leaf-outline" size={26} color="#ffffff" />
            </View>
            <Text className="text-2xl font-bold text-slate-900">Welcome back</Text>
            <Text className="mt-1 text-center text-sm text-slate-500">Sign in to continue your field practicum</Text>
          </View>

          <Controller
            control={control}
            name="email"
            render={({ field: { onChange, onBlur, value } }) => (
              <LabeledInput
                label="Email"
                autoCapitalize="none"
                autoComplete="email"
                keyboardType="email-address"
                placeholder="you@university.edu"
                onBlur={onBlur}
                onChangeText={onChange}
                value={value}
                error={errors.email?.message}
              />
            )}
          />

          <Controller
            control={control}
            name="password"
            render={({ field: { onChange, onBlur, value } }) => (
              <LabeledInput
                label="Password"
                autoCapitalize="none"
                autoComplete="password"
                secureTextEntry
                placeholder="••••••••"
                onBlur={onBlur}
                onChangeText={onChange}
                value={value}
                error={errors.password?.message}
              />
            )}
          />

          {mutation.isError ? (
            <View className="mb-4 flex-row items-center gap-2 rounded-xl bg-rose-50 px-4 py-3">
              <Ionicons name="alert-circle-outline" size={16} color="#e11d48" />
              <Text className="flex-1 text-sm text-rose-600">
                {mutation.error instanceof ApiError && mutation.error.status === 401
                  ? 'Incorrect email or password.'
                  : 'Something went wrong. Please try again.'}
              </Text>
            </View>
          ) : null}

          <PrimaryButton
            label="Sign In"
            onPress={onSubmit}
            disabled={mutation.isPending}
            loading={mutation.isPending}
          />

          <Link href="/(auth)/register" className="mt-6 text-center text-sm font-medium text-brand-600">
            Don&apos;t have an account? Register
          </Link>
        </View>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}
