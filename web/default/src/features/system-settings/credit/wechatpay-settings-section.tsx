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
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import {
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchItem,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import {
  getOptionValue,
  useSystemOptions,
} from '../hooks/use-system-options'
import { useUpdateOption } from '../hooks/use-update-option'

const wechatPayDefaults = {
  'wechatpay_setting.enabled': false,
  'wechatpay_setting.app_id': '',
  'wechatpay_setting.mch_id': '',
  'wechatpay_setting.api_v3_key': '',
  'wechatpay_setting.mch_cert_serial_no': '',
  'wechatpay_setting.mch_private_key': '',
  'wechatpay_setting.platform_public_key': '',
  'wechatpay_setting.platform_cert_serial_no': '',
  'wechatpay_setting.notify_url': '',
}

type WeChatPayValues = typeof wechatPayDefaults

export function WeChatPaySettingsSection() {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const { data, isLoading } = useSystemOptions()

  const current = useMemo(
    () => getOptionValue(data?.data, wechatPayDefaults),
    [data?.data]
  )

  const form = useForm<WeChatPayValues>({
    values: current,
  })

  const { isDirty, isSubmitting } = form.formState
  const enabled = form.watch('wechatpay_setting.enabled')

  async function onSubmit(values: WeChatPayValues) {
    const updates: Array<{ key: string; value: string }> = []
    for (const key of Object.keys(values) as Array<keyof WeChatPayValues>) {
      const next = values[key]
      const prev = current[key]
      if (next !== prev) {
        updates.push({ key: String(key), value: String(next) })
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

  const textField = (
    key: keyof WeChatPayValues,
    label: string,
    opts?: { password?: boolean; textarea?: boolean; description?: string }
  ) => (
    <FormField
      control={form.control}
      name={key}
      render={({ field }) => (
        <FormItem>
          <FormLabel>{t(label)}</FormLabel>
          <FormControl>
            {opts?.textarea ? (
              <Textarea
                rows={4}
                autoComplete='off'
                {...field}
                value={String(field.value ?? '')}
              />
            ) : (
              <Input
                type={opts?.password ? 'password' : 'text'}
                autoComplete='off'
                {...field}
                value={String(field.value ?? '')}
              />
            )}
          </FormControl>
          {opts?.description ? (
            <FormDescription>{t(opts.description)}</FormDescription>
          ) : null}
        </FormItem>
      )}
    />
  )

  return (
    <SettingsSection title={t('WeChat Pay')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)} autoComplete='off'>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending || isSubmitting || isLoading}
            isSaveDisabled={!isDirty}
            saveLabel='Save WeChat Pay settings'
          />

          <FormField
            control={form.control}
            name='wechatpay_setting.enabled'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>{t('Enable WeChat Pay')}</FormLabel>
                  <FormDescription>
                    {t('Enable WeChat Pay (JSAPI / Mini Program) for top-up')}
                  </FormDescription>
                </SettingsSwitchContent>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={(v) => field.onChange(v)}
                  />
                </FormControl>
              </SettingsSwitchItem>
            )}
          />

          {enabled ? (
            <>
              {textField('wechatpay_setting.app_id', 'Mini Program AppID')}
              {textField('wechatpay_setting.mch_id', 'Merchant ID (mchid)')}
              {textField('wechatpay_setting.api_v3_key', 'APIv3 Key', {
                password: true,
                description: 'Used to decrypt payment callbacks (32 bytes)',
              })}
              {textField(
                'wechatpay_setting.mch_cert_serial_no',
                'Merchant Certificate Serial No.'
              )}
              {textField('wechatpay_setting.mch_private_key', 'Merchant Private Key (PEM)', {
                textarea: true,
                description: 'PKCS#8 RSA private key for request signing',
              })}
              {textField(
                'wechatpay_setting.platform_public_key',
                'Platform Public Key / Certificate (PEM)',
                { textarea: true, description: 'Used to verify callback signature' }
              )}
              {textField(
                'wechatpay_setting.platform_cert_serial_no',
                'Platform Certificate Serial No.'
              )}
              {textField('wechatpay_setting.notify_url', 'Payment Notify URL', {
                description: 'WeChat will POST payment results to this URL',
              })}
            </>
          ) : null}
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
