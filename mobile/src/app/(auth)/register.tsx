import { zodResolver } from '@hookform/resolvers/zod';
import { Ionicons } from '@expo/vector-icons';
import { useMutation, useQuery } from '@tanstack/react-query';
import { Link, router } from 'expo-router';
import { Controller, useForm } from 'react-hook-form';
import { ScrollView, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { FormField } from '@/components/ui/form-field';
import { LabeledInput } from '@/components/ui/labeled-input';
import { LoadingState } from '@/components/ui/loading-state';
import { PillSelect } from '@/components/ui/pill-select';
import { PrimaryButton } from '@/components/ui/primary-button';
import { registerSchema, type RegisterFormValues } from '@/features/auth/schemas';
import { ApiError } from '@/lib/api-client';
import * as authService from '@/services/auth';
import * as directoryService from '@/services/directory';
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
    watch,
    setValue,
    formState: { errors },
  } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: { fullName: '', email: '', password: '', role: 'student', institutionId: '', agencyId: '' },
  });

  const role = watch('role');
  const needsInstitution = role === 'student' || role === 'faculty_supervisor';
  const needsAgency = role === 'agency_supervisor';

  const { data: institutions, isLoading: institutionsLoading } = useQuery({
    queryKey: ['public', 'institutions'],
    queryFn: directoryService.listPublicInstitutions,
    enabled: needsInstitution,
  });
  const { data: agencies, isLoading: agenciesLoading } = useQuery({
    queryKey: ['public', 'agencies'],
    queryFn: directoryService.listPublicAgencies,
    enabled: needsAgency,
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
    <SafeAreaView className="flex-1 bg-white">
      <ScrollView className="flex-1" contentContainerClassName="px-7 py-10">
        <View className="mb-8 items-center">
          <View className="mb-4 h-14 w-14 items-center justify-center rounded-2xl bg-brand-600">
            <Ionicons name="leaf-outline" size={26} color="#ffffff" />
          </View>
          <Text className="text-2xl font-bold text-slate-900">Create your account</Text>
          <Text className="mt-1 text-center text-sm text-slate-500">Register as a Student or Supervisor</Text>
        </View>

        <Controller
          control={control}
          name="fullName"
          render={({ field: { onChange, onBlur, value } }) => (
            <LabeledInput
              label="Full name"
              placeholder="Jane Doe"
              onBlur={onBlur}
              onChangeText={onChange}
              value={value}
              error={errors.fullName?.message}
            />
          )}
        />

        <Controller
          control={control}
          name="email"
          render={({ field: { onChange, onBlur, value } }) => (
            <LabeledInput
              label="Email"
              autoCapitalize="none"
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
              secureTextEntry
              placeholder="At least 8 characters"
              onBlur={onBlur}
              onChangeText={onChange}
              value={value}
              error={errors.password?.message}
            />
          )}
        />

        <FormField label="I am a">
          <Controller
            control={control}
            name="role"
            render={({ field: { onChange, value } }) => (
              <PillSelect
                options={ROLE_OPTIONS}
                value={value}
                onChange={(next) => {
                  onChange(next);
                  setValue('institutionId', '');
                  setValue('agencyId', '');
                }}
              />
            )}
          />
        </FormField>

        {needsInstitution ? (
          <FormField label="University" error={errors.institutionId?.message}>
            {institutionsLoading ? (
              <LoadingState compact />
            ) : (
              <Controller
                control={control}
                name="institutionId"
                render={({ field: { onChange, value } }) => (
                  <PillSelect
                    options={(institutions ?? []).map((inst) => ({ value: inst.id, label: inst.name }))}
                    value={value ?? ''}
                    onChange={onChange}
                  />
                )}
              />
            )}
          </FormField>
        ) : null}

        {needsAgency ? (
          <FormField label="Agency" error={errors.agencyId?.message}>
            {agenciesLoading ? (
              <LoadingState compact />
            ) : (
              <Controller
                control={control}
                name="agencyId"
                render={({ field: { onChange, value } }) => (
                  <PillSelect
                    options={(agencies ?? []).map((agency) => ({ value: agency.id, label: agency.name }))}
                    value={value ?? ''}
                    onChange={onChange}
                  />
                )}
              />
            )}
          </FormField>
        ) : null}

        {mutation.isError ? (
          <View className="mb-4 flex-row items-center gap-2 rounded-xl bg-rose-50 px-4 py-3">
            <Ionicons name="alert-circle-outline" size={16} color="#e11d48" />
            <Text className="flex-1 text-sm text-rose-600">
              {mutation.error instanceof ApiError && mutation.error.status === 409
                ? 'That email is already registered.'
                : 'Something went wrong. Please try again.'}
            </Text>
          </View>
        ) : null}

        <PrimaryButton
          label="Register"
          onPress={onSubmit}
          disabled={mutation.isPending}
          loading={mutation.isPending}
        />

        <Link href="/(auth)/login" className="mt-6 text-center text-sm font-medium text-brand-600">
          Already have an account? Sign in
        </Link>
      </ScrollView>
    </SafeAreaView>
  );
}
