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
import { Dialog } from '@/components/dialog'
import {
  ModelPricingEditorPanel,
  type ModelPricingEditorPanelHandle,
  type ModelRatioData,
} from '@/features/system-settings/models/model-pricing-sheet'
import { createModelPriceSchedule } from './api'

interface ModelPriceScheduleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ModelPriceScheduleDialog(props: ModelPriceScheduleDialogProps) {
  // 仅在打开时挂载，使内部 state 从空白初始化，无需 effect 重置
  if (!props.open) return null
  return <ModelPriceScheduleDialogInner {...props} />
}

function ModelPriceScheduleDialogInner(props: ModelPriceScheduleDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const panelRef = React.useRef<ModelPricingEditorPanelHandle>(null)
  const [effectiveAt, setEffectiveAt] = React.useState('')

  const mutation = useMutation({
    mutationFn: createModelPriceSchedule,
    onSuccess: () => {
      toast.success(t('Saved successfully'))
      queryClient.invalidateQueries({ queryKey: ['model-price-schedules'] })
      props.onOpenChange(false)
    },
  })

  const handleSubmit = async () => {
    if (!effectiveAt) {
      toast.error(t('Effective time is required'))
      return
    }
    const ts = Math.floor(new Date(effectiveAt).getTime() / 1000)
    if (ts <= Math.floor(Date.now() / 1000)) {
      toast.error(t('Effective time must be in the future'))
      return
    }
    const data: ModelRatioData | null =
      (await panelRef.current?.commitDraft()) ?? null
    if (!data) {
      // 价格表单校验失败：字段上已显示错误，再给一条整体提示
      toast.error(t('Please complete the model pricing form'))
      return
    }
    mutation.mutate({
      model_name: data.name,
      payload: {
        name: data.name,
        billingMode: data.billingMode ?? 'per-token',
        price: data.price ?? '',
        ratio: data.ratio ?? '',
        cacheRatio: data.cacheRatio ?? '',
        createCacheRatio: data.createCacheRatio ?? '',
        completionRatio: data.completionRatio ?? '',
        imageRatio: data.imageRatio ?? '',
        audioRatio: data.audioRatio ?? '',
        audioCompletionRatio: data.audioCompletionRatio ?? '',
        billingExpr: data.billingExpr ?? '',
        requestRuleExpr: data.requestRuleExpr ?? '',
      },
      effective_at: ts,
    })
  }

  return (
    <Dialog
      open={props.open}
      onOpenChange={props.onOpenChange}
      title={t('Add Schedule')}
      contentClassName='sm:max-w-2xl xl:max-w-5xl'
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
          <Label>{t('Effective time')}</Label>
          <Input
            type='datetime-local'
            value={effectiveAt}
            onChange={(e) => setEffectiveAt(e.target.value)}
          />
        </div>
        <ModelPricingEditorPanel
          ref={panelRef}
          className='h-[min(60vh,560px)] min-h-0'
        />
      </div>
    </Dialog>
  )
}
