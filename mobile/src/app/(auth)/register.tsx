import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery } from '@tanstack/react-query';
import { Link, router } from 'expo-router';
import { Controller, useForm } from 'react-hook-form';
import { ActivityIndicator, Pressable, ScrollView, Text, View } from 'react-native';

import { LabeledInput } from '@/components/ui/labeled-input';
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
    <ScrollView className="flex-1 bg-white" contentContainerClassName="justify-center px-6 py-12">
      <Text className="mb-1 text-2xl font-semibold text-slate-900">Create your account</Text>
      <Text className="mb-8 text-slate-500">Register as a Student or Supervisor.</Text>

      <Controller
        control={control}
        name="fullName"
        render={({ field: { onChange, onBlur, value } }) => (
          <LabeledInput label="Full name" placeholder="Jane Doe" onBlur={onBlur} onChangeText={onChange} value={value} error={errors.fullName?.message} />
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

      <Text className="mb-2 mt-3 text-sm font-medium text-slate-700">I am a</Text>
      <Controller
        control={control}
        name="role"
        render={({ field: { onChange, value } }) => (
          <View className="mb-4 flex-row flex-wrap gap-2">
            {ROLE_OPTIONS.map((option) => (
              <Pressable
                key={option.value}
                onPress={() => {
                  onChange(option.value);
                  setValue('institutionId', '');
                  setValue('agencyId', '');
                }}
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

      {needsInstitution ? (
        <View className="mb-3">
          <Text className="mb-1 text-sm font-medium text-slate-700">University</Text>
          {institutionsLoading ? (
            <ActivityIndicator />
          ) : (
            <Controller
              control={control}
              name="institutionId"
              render={({ field: { onChange, value } }) => (
                <View className="flex-row flex-wrap gap-2">
                  {(institutions ?? []).map((inst) => (
                    <Pressable
                      key={inst.id}
                      onPress={() => onChange(inst.id)}
                      className={`rounded-full border px-4 py-2 ${
                        value === inst.id ? 'border-brand-600 bg-brand-600' : 'border-slate-300 bg-white'
                      }`}
                    >
                      <Text className={value === inst.id ? 'text-white' : 'text-slate-700'}>{inst.name}</Text>
                    </Pressable>
                  ))}
                </View>
              )}
            />
          )}
          {errors.institutionId ? <Text className="mt-1 text-sm text-red-600">{errors.institutionId.message}</Text> : null}
        </View>
      ) : null}

      {needsAgency ? (
        <View className="mb-3">
          <Text className="mb-1 text-sm font-medium text-slate-700">Agency</Text>
          {agenciesLoading ? (
            <ActivityIndicator />
          ) : (
            <Controller
              control={control}
              name="agencyId"
              render={({ field: { onChange, value } }) => (
                <View className="flex-row flex-wrap gap-2">
                  {(agencies ?? []).map((agency) => (
                    <Pressable
                      key={agency.id}
                      onPress={() => onChange(agency.id)}
                      className={`rounded-full border px-4 py-2 ${
                        value === agency.id ? 'border-brand-600 bg-brand-600' : 'border-slate-300 bg-white'
                      }`}
                    >
                      <Text className={value === agency.id ? 'text-white' : 'text-slate-700'}>{agency.name}</Text>
                    </Pressable>
                  ))}
                </View>
              )}
            />
          )}
          {errors.agencyId ? <Text className="mt-1 text-sm text-red-600">{errors.agencyId.message}</Text> : null}
        </View>
      ) : null}

      {mutation.isError ? (
        <Text className="mb-2 text-sm text-red-600">
          {mutation.error instanceof ApiError && mutation.error.status === 409
            ? 'That email is already registered.'
            : 'Something went wrong. Please try again.'}
        </Text>
      ) : null}

      <PrimaryButton label={mutation.isPending ? 'Creating account…' : 'Register'} onPress={onSubmit} disabled={mutation.isPending} />

      <Link href="/(auth)/login" className="mt-6 text-center text-brand-600">
        Already have an account? Sign in
      </Link>
    </ScrollView>
  );
}
