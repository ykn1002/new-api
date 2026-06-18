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
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { getCurrencyDisplay, getCurrencyLabel } from '@/lib/currency'
import { formatQuota, parseQuotaFromDollars } from '@/lib/format'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Dialog } from '@/components/dialog'
import { listRechargeTiers } from '@/features/system-settings/credit/api'
import { adjustUserQuota } from '../api'
import type { QuotaAdjustMode } from '../types'

interface UserQuotaDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  userId: number
  currentQuota: number
  onSuccess: () => void
}

export function UserQuotaDialog(props: UserQuotaDialogProps) {
  const { t } = useTranslation()
  const [mode, setMode] = useState<QuotaAdjustMode>('add')
  const [amount, setAmount] = useState('')
  const [loading, setLoading] = useState(false)
  const [tierId, setTierId] = useState('')
  const [giftAmount, setGiftAmount] = useState('')
  const [reason, setReason] = useState('')

  const { data: tiers } = useQuery({
    queryKey: ['recharge-tiers'],
    queryFn: listRechargeTiers,
    enabled: props.open,
  })

  const { meta: currencyMeta } = getCurrencyDisplay()
  const currencyLabel = getCurrencyLabel()
  const tokensOnly = currencyMeta.kind === 'tokens'

  const amountValue = parseFloat(amount) || 0
  const quotaValue = parseQuotaFromDollars(Math.abs(amountValue))

  const getPreviewText = () => {
    const current = props.currentQuota
    const val = quotaValue
    switch (mode) {
      case 'add':
        return `${t('Current quota')}: ${formatQuota(current)}  +${formatQuota(val)} = ${formatQuota(current + val)}`
      case 'subtract':
        return `${t('Current quota')}: ${formatQuota(current)}  -${formatQuota(val)} = ${formatQuota(current - val)}`
      case 'override': {
        const overrideQuota = parseQuotaFromDollars(amountValue)
        return `${t('Current quota')}: ${formatQuota(current)} → ${formatQuota(overrideQuota)}`
      }
      default:
        return ''
    }
  }

  const handleConfirm = async () => {
    if (mode === 'tier') {
      if (!tierId) {
        toast.error(t('Please select a recharge tier'))
        return
      }
      setLoading(true)
      try {
        const result = await adjustUserQuota({
          id: props.userId,
          action: 'add_quota',
          mode: 'tier',
          value: 0,
          tier_id: tierId,
          gift_value: parseQuotaFromDollars(parseFloat(giftAmount) || 0),
          reason: reason || undefined,
        })
        if (result.success) {
          toast.success(t('Quota adjusted successfully'))
          resetForm()
          props.onOpenChange(false)
          props.onSuccess()
        } else {
          toast.error(result.message || t('Failed to adjust quota'))
        }
      } catch (e: unknown) {
        toast.error(e instanceof Error ? e.message : t('Failed to adjust quota'))
      } finally {
        setLoading(false)
      }
      return
    }

    if (!amount && mode !== 'override') return
    if (quotaValue <= 0 && mode !== 'override') return

    setLoading(true)
    try {
      const value =
        mode === 'override' ? parseQuotaFromDollars(amountValue) : quotaValue
      const result = await adjustUserQuota({
        id: props.userId,
        action: 'add_quota',
        mode,
        value: mode === 'override' ? value : Math.abs(value),
      })
      if (result.success) {
        toast.success(t('Quota adjusted successfully'))
        resetForm()
        props.onOpenChange(false)
        props.onSuccess()
      } else {
        toast.error(result.message || t('Failed to adjust quota'))
      }
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : t('Failed to adjust quota'))
    } finally {
      setLoading(false)
    }
  }

  const resetForm = () => {
    setAmount('')
    setMode('add')
    setTierId('')
    setGiftAmount('')
    setReason('')
  }

  const handleCancel = () => {
    resetForm()
    props.onOpenChange(false)
  }

  const placeholder = tokensOnly
    ? t('Enter amount in tokens')
    : t('Enter amount in {{currency}}', { currency: currencyLabel })

  return (
    <Dialog
      open={props.open}
      onOpenChange={props.onOpenChange}
      title={t('Adjust Quota')}
      description={t('Select an operation mode and enter the amount')}
      contentHeight='auto'
      bodyClassName='space-y-4'
      footer={
        <>
          <Button variant='outline' onClick={handleCancel}>
            {t('Cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={loading}>
            {loading ? t('Processing...') : t('Confirm')}
          </Button>
        </>
      }
    >
      <div className='space-y-4'>
        {mode !== 'tier' && (
          <div className='text-muted-foreground text-sm'>{getPreviewText()}</div>
        )}

        <div className='space-y-2'>
          <Label>{t('Mode')}</Label>
          <div className='flex flex-wrap gap-1'>
            {(['add', 'subtract', 'override', 'tier'] as const).map((m) => (
              <Button
                key={m}
                type='button'
                variant='outline'
                size='sm'
                className={cn(
                  mode === m &&
                    'bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground'
                )}
                onClick={() => {
                  setMode(m)
                  setAmount('')
                }}
              >
                {m === 'add'
                  ? t('Add')
                  : m === 'subtract'
                    ? t('Subtract')
                    : m === 'override'
                      ? t('Override')
                      : t('By Tier')}
              </Button>
            ))}
          </div>
        </div>

        {mode === 'tier' ? (
          <>
            <div className='space-y-2'>
              <Label>{t('Recharge Tier')}</Label>
              <Select
                value={tierId}
                onValueChange={(value) => setTierId(value ?? '')}
              >
                <SelectTrigger>
                  <SelectValue placeholder={t('Select a recharge tier')} />
                </SelectTrigger>
                <SelectContent alignItemWithTrigger={false}>
                  {(tiers ?? [])
                    .filter((tier) => tier.status === 1)
                    .map((tier) => (
                      <SelectItem key={tier.id} value={tier.id}>
                        {tier.name} · ¥{tier.amount} · {tier.base_credits}
                        {tier.gift_credits > 0
                          ? ` +${tier.gift_credits}`
                          : ''}
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
            </div>
            <div className='space-y-2'>
              <Label>
                {t('Extra gift')} ({currencyLabel})
              </Label>
              <Input
                type='number'
                min={0}
                placeholder={t('Optional extra gift amount')}
                value={giftAmount}
                onChange={(e) => setGiftAmount(e.target.value)}
              />
            </div>
            <div className='space-y-2'>
              <Label>{t('Reason')}</Label>
              <Textarea
                rows={2}
                placeholder={t('Optional note for audit log')}
                value={reason}
                onChange={(e) => setReason(e.target.value)}
              />
            </div>
          </>
        ) : (
          <div className='space-y-2'>
            <Label>
              {t('Amount')} ({currencyLabel})
            </Label>
            <Input
              type='number'
              step={tokensOnly ? 1 : 0.000001}
              min={mode === 'override' ? undefined : 0}
              placeholder={placeholder}
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleConfirm()
              }}
            />
          </div>
        )}
      </div>
    </Dialog>
  )
}
