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
import * as React from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Dialog } from '@/components/dialog'
import { saveRechargeTier, type RechargeTier } from './api'

interface RechargeTierDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  tier: RechargeTier | null
}

const emptyForm: RechargeTier = {
  id: '',
  name: '',
  amount: 0,
  base_credits: 0,
  gift_credits: 0,
  validity_days: 0,
  status: 1,
  sort_order: 0,
}

export function RechargeTierDialog(props: RechargeTierDialogProps) {
  // 仅在打开时挂载内部表单，使初始 state 直接来自 props，无需 effect 重置
  if (!props.open) return null
  return (
    <RechargeTierDialogInner
      key={props.tier?.id ?? 'new'}
      {...props}
    />
  )
}

function RechargeTierDialogInner(props: RechargeTierDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [form, setForm] = React.useState<RechargeTier>(
    props.tier ? { ...props.tier } : { ...emptyForm }
  )

  const mutation = useMutation({
    mutationFn: saveRechargeTier,
    onSuccess: () => {
      toast.success(t('Saved successfully'))
      queryClient.invalidateQueries({ queryKey: ['recharge-tiers'] })
      props.onOpenChange(false)
    },
  })

  const setField = <K extends keyof RechargeTier>(
    key: K,
    value: RechargeTier[K]
  ) => setForm((prev) => ({ ...prev, [key]: value }))

  // 数字输入：编辑期允许清空（用 NaN 表示空），提交时再兜底为 0；展示时 NaN 显示为空串
  const numberValue = (v: number) => (Number.isNaN(v) ? '' : v)
  const numberField = (key: keyof RechargeTier) => (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    const raw = e.target.value
    setField(key, (raw === '' ? NaN : Number(raw)) as never)
  }
  const num = (v: number) => (Number.isFinite(v) ? v : 0)

  const handleSubmit = () => {
    if (!form.name.trim()) {
      toast.error(t('Tier name is required'))
      return
    }
    if (num(form.amount) <= 0) {
      toast.error(t('Amount must be greater than 0'))
      return
    }
    mutation.mutate({
      ...form,
      amount: num(form.amount),
      base_credits: num(form.base_credits),
      gift_credits: num(form.gift_credits),
      validity_days: num(form.validity_days),
      sort_order: num(form.sort_order),
    })
  }

  return (
    <Dialog
      open={props.open}
      onOpenChange={props.onOpenChange}
      title={props.tier ? t('Edit Tier') : t('Add Tier')}
      footer={
        <div className='flex justify-end gap-2'>
          <Button variant='outline' onClick={() => props.onOpenChange(false)}>
            {t('Cancel')}
          </Button>
          <Button onClick={handleSubmit} disabled={mutation.isPending}>
            {t('Save')}
          </Button>
        </div>
      }
    >
      <div className='grid gap-4'>
        <div className='grid gap-2'>
          <Label>{t('Tier Name')}</Label>
          <Input
            value={form.name}
            onChange={(e) => setField('name', e.target.value)}
            placeholder={t('e.g. Standard / Pro / Flagship')}
          />
        </div>
        <div className='grid grid-cols-2 gap-4'>
          <div className='grid gap-2'>
            <Label>{t('Amount (CNY)')}</Label>
            <Input
              type='number'
              value={numberValue(form.amount)}
              onChange={numberField('amount')}
            />
          </div>
          <div className='grid gap-2'>
            <Label>{t('Validity (days, 0 = permanent)')}</Label>
            <Input
              type='number'
              value={numberValue(form.validity_days)}
              onChange={numberField('validity_days')}
            />
          </div>
        </div>
        <div className='grid grid-cols-2 gap-4'>
          <div className='grid gap-2'>
            <Label>{t('Base Credits')}</Label>
            <Input
              type='number'
              value={numberValue(form.base_credits)}
              onChange={numberField('base_credits')}
            />
          </div>
          <div className='grid gap-2'>
            <Label>{t('Gift Credits')}</Label>
            <Input
              type='number'
              value={numberValue(form.gift_credits)}
              onChange={numberField('gift_credits')}
            />
          </div>
        </div>
        <div className='grid grid-cols-2 gap-4'>
          <div className='grid gap-2'>
            <Label>{t('Sort Order')}</Label>
            <Input
              type='number'
              value={numberValue(form.sort_order)}
              onChange={numberField('sort_order')}
            />
          </div>
          <div className='flex items-center justify-between'>
            <Label>{t('On shelf')}</Label>
            <Switch
              checked={form.status === 1}
              onCheckedChange={(checked) =>
                setField('status', checked ? 1 : 0)
              }
            />
          </div>
        </div>
      </div>
    </Dialog>
  )
}
