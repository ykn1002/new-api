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
  const [modelName, setModelName] = React.useState('')
  const [modelRatio, setModelRatio] = React.useState('')
  const [completionRatio, setCompletionRatio] = React.useState('')
  const [effectiveAt, setEffectiveAt] = React.useState('')

  const mutation = useMutation({
    mutationFn: createModelPriceSchedule,
    onSuccess: () => {
      toast.success(t('Saved successfully'))
      queryClient.invalidateQueries({ queryKey: ['model-price-schedules'] })
      props.onOpenChange(false)
    },
  })

  const handleSubmit = () => {
    if (!modelName.trim()) {
      toast.error(t('Model name is required'))
      return
    }
    const mr = Number(modelRatio) || 0
    const cr = Number(completionRatio) || 0
    if (mr <= 0 && cr <= 0) {
      toast.error(t('Configure at least one of model ratio or completion ratio'))
      return
    }
    if (!effectiveAt) {
      toast.error(t('Effective time is required'))
      return
    }
    const ts = Math.floor(new Date(effectiveAt).getTime() / 1000)
    if (ts <= Math.floor(Date.now() / 1000)) {
      toast.error(t('Effective time must be in the future'))
      return
    }
    mutation.mutate({
      model_name: modelName.trim(),
      model_ratio: mr,
      completion_ratio: cr,
      effective_at: ts,
    })
  }

  return (
    <Dialog
      open={props.open}
      onOpenChange={props.onOpenChange}
      title={t('Add Schedule')}
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
          <Label>{t('Model Name')}</Label>
          <Input
            value={modelName}
            onChange={(e) => setModelName(e.target.value)}
            placeholder='gpt-4o'
          />
        </div>
        <div className='grid grid-cols-2 gap-4'>
          <div className='grid gap-2'>
            <Label>{t('Model ratio')}</Label>
            <Input
              type='number'
              value={modelRatio}
              onChange={(e) => setModelRatio(e.target.value)}
            />
          </div>
          <div className='grid gap-2'>
            <Label>{t('Completion ratio')}</Label>
            <Input
              type='number'
              value={completionRatio}
              onChange={(e) => setCompletionRatio(e.target.value)}
            />
          </div>
        </div>
        <div className='grid gap-2'>
          <Label>{t('Effective time')}</Label>
          <Input
            type='datetime-local'
            value={effectiveAt}
            onChange={(e) => setEffectiveAt(e.target.value)}
          />
        </div>
      </div>
    </Dialog>
  )
}
