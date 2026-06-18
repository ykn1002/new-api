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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { formatTimestampToDate } from '@/lib/format'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { SettingsSection } from '../components/settings-section'
import {
  deleteModelPriceSchedule,
  listModelPriceSchedules,
  type ModelPriceSchedule,
} from './api'
import { ModelPriceScheduleDialog } from './model-price-schedule-dialog'

export function ModelPriceScheduleSection() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [deleteTarget, setDeleteTarget] =
    React.useState<ModelPriceSchedule | null>(null)

  const { data: schedules, isLoading } = useQuery({
    queryKey: ['model-price-schedules'],
    queryFn: () => listModelPriceSchedules(),
  })

  const deleteMutation = useMutation({
    mutationFn: deleteModelPriceSchedule,
    onSuccess: () => {
      toast.success(t('Deleted successfully'))
      queryClient.invalidateQueries({ queryKey: ['model-price-schedules'] })
      setDeleteTarget(null)
    },
  })

  return (
    <SettingsSection title={t('Scheduled Model Pricing')}>
      <p className='text-muted-foreground -mt-2 text-sm'>
        {t(
          'Schedule model ratio changes to take effect at a future time. Only affects future requests.'
        )}
      </p>
      <div className='flex justify-end'>
        <Button size='sm' onClick={() => setDialogOpen(true)}>
          <Plus className='size-4' />
          {t('Add Schedule')}
        </Button>
      </div>

      {isLoading ? (
        <div className='mt-4 space-y-2'>
          {Array.from({ length: 2 }).map((_, i) => (
            <Skeleton key={i} className='h-14 w-full' />
          ))}
        </div>
      ) : !schedules || schedules.length === 0 ? (
        <div className='text-muted-foreground mt-6 text-center text-sm'>
          {t('No scheduled pricing changes.')}
        </div>
      ) : (
        <div className='mt-4 space-y-2'>
          {schedules.map((s) => (
            <div
              key={s.id}
              className='flex items-center justify-between rounded-lg border p-3'
            >
              <div className='min-w-0'>
                <div className='flex items-center gap-2'>
                  <span className='truncate font-medium'>{s.model_name}</span>
                  <Badge variant={s.applied ? 'secondary' : 'default'}>
                    {s.applied ? t('Applied') : t('Pending')}
                  </Badge>
                </div>
                <div className='text-muted-foreground mt-1 text-xs'>
                  {t('Model ratio')} {s.model_ratio} · {t('Completion ratio')}{' '}
                  {s.completion_ratio} · {t('Effective at')}{' '}
                  {formatTimestampToDate(s.effective_at)}
                </div>
              </div>
              {!s.applied && (
                <Button
                  size='icon'
                  variant='ghost'
                  onClick={() => setDeleteTarget(s)}
                >
                  <Trash2 className='text-destructive size-4' />
                </Button>
              )}
            </div>
          ))}
        </div>
      )}

      <ModelPriceScheduleDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
      />

      <ConfirmDialog
        open={deleteTarget != null}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title={t('Delete schedule')}
        desc={t('Are you sure you want to delete this scheduled change?')}
        destructive
        handleConfirm={() => {
          if (deleteTarget) deleteMutation.mutate(deleteTarget.id)
        }}
        isLoading={deleteMutation.isPending}
      />
    </SettingsSection>
  )
}
