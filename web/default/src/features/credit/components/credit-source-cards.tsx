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
import { Gift, Package, Wallet } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatQuota, formatTimestampToDate } from '@/lib/format'
import { Skeleton } from '@/components/ui/skeleton'
import type { CreditSource, CreditSourceView } from '../types'

interface CreditSourceCardsProps {
  sources: CreditSourceView[]
  total: number
  loading?: boolean
}

const SOURCE_META: Record<
  CreditSource,
  { labelKey: string; icon: typeof Gift; color: string }
> = {
  gift: { labelKey: 'Gift credits', icon: Gift, color: 'bg-pink-500' },
  subscription: {
    labelKey: 'Subscription credits',
    icon: Package,
    color: 'bg-blue-500',
  },
  topup: { labelKey: 'Top-up credits', icon: Wallet, color: 'bg-emerald-500' },
}

const SOURCE_ORDER: CreditSource[] = ['gift', 'subscription', 'topup']

export function CreditSourceCards(props: CreditSourceCardsProps) {
  const { t } = useTranslation()

  if (props.loading) {
    return (
      <div className='grid gap-4 md:grid-cols-3'>
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className='rounded-lg border p-4'>
            <Skeleton className='h-4 w-24' />
            <Skeleton className='mt-3 h-7 w-32' />
            <Skeleton className='mt-2 h-3.5 w-28' />
          </div>
        ))}
      </div>
    )
  }

  const bySource = new Map<string, CreditSourceView>()
  for (const s of props.sources) bySource.set(s.source, s)

  return (
    <div className='space-y-4'>
      {/* 堆叠占比条 */}
      <div className='bg-muted flex h-3 w-full overflow-hidden rounded-full'>
        {SOURCE_ORDER.map((src) => {
          const item = bySource.get(src)
          if (!item || props.total <= 0 || item.remaining <= 0) return null
          const pct = (item.remaining / props.total) * 100
          return (
            <div
              key={src}
              className={SOURCE_META[src].color}
              style={{ width: `${pct}%` }}
              title={`${t(SOURCE_META[src].labelKey)} ${pct.toFixed(1)}%`}
            />
          )
        })}
      </div>

      {/* 三来源卡片 */}
      <div className='grid gap-4 md:grid-cols-3'>
        {SOURCE_ORDER.map((src) => {
          const item = bySource.get(src)
          const meta = SOURCE_META[src]
          const remaining = item?.remaining ?? 0
          const pct =
            props.total > 0 ? (remaining / props.total) * 100 : 0
          return (
            <div key={src} className='rounded-lg border p-4'>
              <div className='flex items-center gap-2'>
                <span
                  className={`flex size-7 items-center justify-center rounded-md ${meta.color}/10`}
                >
                  <meta.icon className='size-4' />
                </span>
                <span className='text-muted-foreground text-xs font-medium tracking-wider uppercase'>
                  {t(meta.labelKey)}
                </span>
              </div>
              <div className='text-foreground mt-2 font-mono text-xl font-bold tabular-nums'>
                {formatQuota(remaining)}
              </div>
              <div className='text-muted-foreground/70 mt-1 text-xs'>
                {pct.toFixed(1)}%
                {item && item.nearest_expire > 0 ? (
                  <>
                    {' · '}
                    {t('Valid until')} {formatTimestampToDate(item.nearest_expire)}
                  </>
                ) : (
                  <>
                    {' · '}
                    {t('Long-term valid')}
                  </>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
