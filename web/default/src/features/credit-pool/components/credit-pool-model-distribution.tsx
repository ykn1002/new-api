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
import { useTranslation } from 'react-i18next'
import { formatNumber } from '@/lib/format'
import { Skeleton } from '@/components/ui/skeleton'

interface CreditPoolModelDistributionProps {
  modelCredits: Record<string, number>
  loading?: boolean
}

const BAR_COLORS = [
  'bg-blue-500',
  'bg-emerald-500',
  'bg-amber-500',
  'bg-rose-500',
  'bg-violet-500',
  'bg-cyan-500',
]

export function CreditPoolModelDistribution(
  props: CreditPoolModelDistributionProps
) {
  const { t } = useTranslation()

  const rows = useMemo(() => {
    const entries = Object.entries(props.modelCredits)
      .filter(([, v]) => v > 0)
      .sort((a, b) => b[1] - a[1])
    const total = entries.reduce((sum, [, v]) => sum + v, 0)
    return { entries: entries.slice(0, 8), total }
  }, [props.modelCredits])

  return (
    <div className='rounded-2xl border p-4 sm:p-5'>
      <h3 className='text-base font-semibold'>
        {t('Model credit consumption share')}
      </h3>
      {props.loading ? (
        <div className='mt-4 space-y-3'>
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className='h-6 w-full' />
          ))}
        </div>
      ) : rows.entries.length === 0 ? (
        <div className='text-muted-foreground mt-6 text-center text-sm'>
          {t('No consumption data in range.')}
        </div>
      ) : (
        <div className='mt-4 space-y-3'>
          {rows.entries.map(([model, credits], idx) => {
            const pct = rows.total > 0 ? (credits / rows.total) * 100 : 0
            return (
              <div key={model}>
                <div className='flex items-center justify-between text-sm'>
                  <span className='truncate font-medium'>{model}</span>
                  <span className='text-muted-foreground tabular-nums'>
                    {formatNumber(credits)} · {pct.toFixed(1)}%
                  </span>
                </div>
                <div className='bg-muted mt-1 h-2 w-full overflow-hidden rounded-full'>
                  <div
                    className={BAR_COLORS[idx % BAR_COLORS.length]}
                    style={{ width: `${pct}%`, height: '100%' }}
                  />
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
