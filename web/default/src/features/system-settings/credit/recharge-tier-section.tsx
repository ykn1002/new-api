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
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { SettingsSection } from '../components/settings-section'
import {
  deleteRechargeTier,
  listRechargeTiers,
  type RechargeTier,
} from './api'
import { RechargeTierDialog } from './recharge-tier-dialog'

export function RechargeTierSection() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [editing, setEditing] = React.useState<RechargeTier | null>(null)
  const [deleteTarget, setDeleteTarget] = React.useState<RechargeTier | null>(
    null
  )

  const { data: tiers, isLoading } = useQuery({
    queryKey: ['recharge-tiers'],
    queryFn: listRechargeTiers,
  })

  const deleteMutation = useMutation({
    mutationFn: deleteRechargeTier,
    onSuccess: () => {
      toast.success(t('Deleted successfully'))
      queryClient.invalidateQueries({ queryKey: ['recharge-tiers'] })
      setDeleteTarget(null)
    },
  })

  const openCreate = () => {
    setEditing(null)
    setDialogOpen(true)
  }
  const openEdit = (tier: RechargeTier) => {
    setEditing(tier)
    setDialogOpen(true)
  }

  return (
    <SettingsSection title={t('Recharge Tiers')}>
      <p className='text-muted-foreground -mt-2 text-sm'>
        {t(
          'Configure recharge tiers with base credits, gift credits and validity.'
        )}
      </p>
      <div className='flex justify-end'>
        <Button size='sm' onClick={openCreate}>
          <Plus className='size-4' />
          {t('Add Tier')}
        </Button>
      </div>

      {isLoading ? (
        <div className='mt-4 space-y-2'>
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className='h-14 w-full' />
          ))}
        </div>
      ) : !tiers || tiers.length === 0 ? (
        <div className='text-muted-foreground mt-6 text-center text-sm'>
          {t('No recharge tiers configured yet.')}
        </div>
      ) : (
        <div className='mt-4 space-y-2'>
          {tiers.map((tier) => (
            <div
              key={tier.id}
              className='flex items-center justify-between rounded-lg border p-3'
            >
              <div className='min-w-0'>
                <div className='flex items-center gap-2'>
                  <span className='truncate font-medium'>{tier.name}</span>
                  <Badge variant={tier.status === 1 ? 'default' : 'secondary'}>
                    {tier.status === 1 ? t('On shelf') : t('Off shelf')}
                  </Badge>
                </div>
                <div className='text-muted-foreground mt-1 text-xs'>
                  ¥{tier.amount} · {t('Base')} {tier.base_credits} ·{' '}
                  {t('Gift')} {tier.gift_credits} ·{' '}
                  {tier.validity_days > 0
                    ? t('{{days}} days', { days: tier.validity_days })
                    : t('Permanent')}
                </div>
              </div>
              <div className='flex shrink-0 items-center gap-1'>
                <Button
                  size='icon'
                  variant='ghost'
                  onClick={() => openEdit(tier)}
                >
                  <Pencil className='size-4' />
                </Button>
                <Button
                  size='icon'
                  variant='ghost'
                  onClick={() => setDeleteTarget(tier)}
                >
                  <Trash2 className='text-destructive size-4' />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      <RechargeTierDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        tier={editing}
      />

      <ConfirmDialog
        open={deleteTarget != null}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title={t('Delete recharge tier')}
        desc={t('Are you sure you want to delete "{{name}}"?', {
          name: deleteTarget?.name ?? '',
        })}
        destructive
        handleConfirm={() => {
          if (deleteTarget) deleteMutation.mutate(deleteTarget.id)
        }}
        isLoading={deleteMutation.isPending}
      />
    </SettingsSection>
  )
}
