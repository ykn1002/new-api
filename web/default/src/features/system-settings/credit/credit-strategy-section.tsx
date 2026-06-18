/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useMemo } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
import {
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchItem,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { NumberInput } from '@/components/number-input'
import { SettingsSection } from '../components/settings-section'
import { getOptionValue, useSystemOptions } from '../hooks/use-system-options'
import { useUpdateOption } from '../hooks/use-update-option'
import { numberFieldProps } from '../utils/numeric-field'

const creditStrategyDefaults = {
  'credit_setting.round_to_display_precision': false,
  'credit_setting.low_balance_warn_percent': 20,
  'credit_setting.notify_root_on_low_balance': false,
  'credit_setting.gift_default_validity_days': 30,
  'credit_setting.topup_default_validity_days': 0,
  'credit_setting.expire_scan_interval_minutes': 5,
  'credit_setting.default_model_enabled': false,
  'credit_setting.default_model': '',
  'credit_setting.enforce_model_status': false,
}

type CreditStrategyValues = typeof creditStrategyDefaults

export function CreditStrategySection() {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const { data, isLoading } = useSystemOptions()

  const current = useMemo(
    () => getOptionValue(data?.data, creditStrategyDefaults),
    [data?.data]
  )

  const form = useForm<CreditStrategyValues>({ values: current })
  const { isDirty, isSubmitting } = form.formState
  const defaultModelEnabled = form.watch('credit_setting.default_model_enabled')

  async function onSubmit(values: CreditStrategyValues) {
    const updates: Array<{ key: string; value: string }> = []
    for (const key of Object.keys(values) as Array<keyof CreditStrategyValues>) {
      if (values[key] !== current[key]) {
        updates.push({ key: String(key), value: String(values[key]) })
      }
    }
    if (updates.length === 0) {
      toast.info(t('No changes to save'))
      return
    }
    for (const update of updates) {
      await updateOption.mutateAsync(update)
    }
    form.reset(values)
  }

  const numberField = (
    key: keyof CreditStrategyValues,
    label: string,
    description?: string
  ) => (
    <FormField
      control={form.control}
      name={key}
      render={({ field }) => (
        <FormItem>
          <FormLabel>{t(label)}</FormLabel>
          <FormControl>
            <NumberInput min={0} {...numberFieldProps(field)} />
          </FormControl>
          {description ? (
            <FormDescription>{t(description)}</FormDescription>
          ) : null}
        </FormItem>
      )}
    />
  )

  const switchField = (
    key: keyof CreditStrategyValues,
    label: string,
    description: string
  ) => (
    <FormField
      control={form.control}
      name={key}
      render={({ field }) => (
        <SettingsSwitchItem>
          <SettingsSwitchContent>
            <FormLabel>{t(label)}</FormLabel>
            <FormDescription>{t(description)}</FormDescription>
          </SettingsSwitchContent>
          <FormControl>
            <Switch
              checked={Boolean(field.value)}
              onCheckedChange={(v) => field.onChange(v)}
            />
          </FormControl>
        </SettingsSwitchItem>
      )}
    />
  )

  return (
    <SettingsSection title={t('Global Strategy')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)} autoComplete='off'>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending || isSubmitting || isLoading}
            isSaveDisabled={!isDirty}
            saveLabel='Save global strategy'
          />

          {/* 计费取整 */}
          {switchField(
            'credit_setting.round_to_display_precision',
            'Round settlement to 2-decimal credits',
            'Round each call settlement up to 2-decimal credit precision (CUSTOM currency only)'
          )}

          <Separator />

          {/* 余额预警 */}
          {numberField(
            'credit_setting.low_balance_warn_percent',
            'Low balance warning threshold (%)',
            'Warn when remaining / baseline is below this percent (0 = disabled). Baseline = sum of active batch initial quota.'
          )}
          {switchField(
            'credit_setting.notify_root_on_low_balance',
            'Notify operators on low balance',
            'Also notify the root user when a user runs low on balance'
          )}

          <Separator />

          {/* 有效期默认值 + 扫描周期 */}
          {numberField(
            'credit_setting.gift_default_validity_days',
            'Gift credits default validity (days)',
            '0 = permanent'
          )}
          {numberField(
            'credit_setting.topup_default_validity_days',
            'Top-up credits default validity (days)',
            '0 = permanent'
          )}
          {numberField(
            'credit_setting.expire_scan_interval_minutes',
            'Expiry scan interval (minutes)',
            'How often the background task expires due batches and applies due price schedules'
          )}

          <Separator />

          {/* 全局默认模型 */}
          {switchField(
            'credit_setting.default_model_enabled',
            'Enable global default model',
            'Ignore the client-specified model and force-route all requests to the default model'
          )}
          {defaultModelEnabled ? (
            <FormField
              control={form.control}
              name='credit_setting.default_model'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Default model')}</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='gpt-4o'
                      value={String(field.value ?? '')}
                      onChange={(e) => field.onChange(e.target.value)}
                    />
                  </FormControl>
                </FormItem>
              )}
            />
          ) : null}
          {switchField(
            'credit_setting.enforce_model_status',
            'Enforce model offline status',
            'Reject calls to models that are offline (Model.Status != enabled)'
          )}
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
