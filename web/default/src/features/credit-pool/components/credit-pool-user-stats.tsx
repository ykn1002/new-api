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
import { useTranslation } from 'react-i18next'
import { formatNumber } from '@/lib/format'
import { Skeleton } from '@/components/ui/skeleton'
import type { UserCreditStats } from '@/features/system-settings/credit/api'

interface CreditPoolUserStatsProps {
  stats?: UserCreditStats
}

export function CreditPoolUserStats(props: CreditPoolUserStatsProps) {
  const { t } = useTranslation()
  const s = props.stats

  const items = [
    { label: t('All'), value: s?.total, tone: 'text-foreground' },
    { label: t('Normal'), value: s?.normal, tone: 'text-success' },
    { label: t('Low balance'), value: s?.low_balance, tone: 'text-warning' },
    { label: t('Exhausted'), value: s?.exhausted, tone: 'text-destructive' },
  ]

  return (
    <div className='rounded-2xl border p-4 sm:p-5'>
      <h3 className='text-base font-semibold'>
        {t('User balance status')}
      </h3>
      <div className='mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4'>
        {items.map((item) => (
          <div key={item.label} className='bg-muted/40 rounded-lg p-3'>
            <div className='text-muted-foreground text-xs font-medium'>
              {item.label}
            </div>
            {item.value === undefined ? (
              <Skeleton className='mt-2 h-7 w-12' />
            ) : (
              <div
                className={`mt-1.5 font-mono text-xl font-bold tabular-nums ${item.tone}`}
              >
                {formatNumber(item.value)}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
