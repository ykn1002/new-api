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
import {
  Coins,
  Download,
  TrendingUp,
  Users,
  Wallet,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatNumber } from '@/lib/format'
import { Button } from '@/components/ui/button'
import { SectionPageLayout } from '@/components/layout'
import { StatCard } from '@/features/dashboard/components/ui/stat-card'
import {
  buildCreditExportUrl,
  getCreditOverview,
  getUserCreditStats,
} from '@/features/system-settings/credit/api'
import { CreditPoolModelDistribution } from './components/credit-pool-model-distribution'
import { CreditPoolUserStats } from './components/credit-pool-user-stats'

type RangeKey = 'today' | 'week' | 'month'

function rangeBounds(range: RangeKey): { start: number; end: number } {
  const end = Math.floor(Date.now() / 1000)
  if (range === 'today') {
    const now = new Date()
    now.setHours(0, 0, 0, 0)
    return { start: Math.floor(now.getTime() / 1000), end }
  }
  const days = range === 'week' ? 7 : 30
  return { start: end - days * 86400, end }
}

export function CreditPool() {
  const { t } = useTranslation()
  const [range, setRange] = useState<RangeKey>('week')

  const { data: overview, isLoading } = useQuery({
    queryKey: ['credit-overview', range],
    queryFn: () => {
      const { start, end } = rangeBounds(range)
      return getCreditOverview(start, end)
    },
  })

  const { data: userStats } = useQuery({
    queryKey: ['credit-user-stats'],
    queryFn: getUserCreditStats,
  })

  const ranges: RangeKey[] = ['today', 'week', 'month']
  const rangeLabels: Record<RangeKey, string> = {
    today: t('Today'),
    week: t('This Week'),
    month: t('This Month'),
  }

  const handleExport = (type: 'model' | 'user') => {
    const { start, end } = rangeBounds(range)
    const url = buildCreditExportUrl(type, start, end)
    window.open(url, '_blank')
  }

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Credit Pool')}</SectionPageLayout.Title>
      <SectionPageLayout.Actions>
        <div className='flex items-center gap-1 rounded-lg border p-0.5'>
          {ranges.map((r) => (
            <Button
              key={r}
              size='sm'
              variant={range === r ? 'secondary' : 'ghost'}
              onClick={() => setRange(r)}
            >
              {rangeLabels[r]}
            </Button>
          ))}
        </div>
        <Button size='sm' variant='outline' onClick={() => handleExport('model')}>
          <Download className='size-4' />
          {t('Export by model')}
        </Button>
        <Button size='sm' variant='outline' onClick={() => handleExport('user')}>
          <Download className='size-4' />
          {t('Export by user')}
        </Button>
      </SectionPageLayout.Actions>
      <SectionPageLayout.Content>
        <div className='space-y-6'>
          {/* 总览卡片 */}
          <div className='grid gap-4 rounded-2xl border p-4 sm:p-5 md:grid-cols-3 xl:grid-cols-5'>
            <StatCard
              title={t('Total Token consumption')}
              value={formatNumber(overview?.total_tokens ?? 0)}
              description={t('Tokens consumed in range')}
              icon={TrendingUp}
              loading={isLoading}
              tone='teal'
            />
            <StatCard
              title={t('Total credit consumption')}
              value={formatNumber(overview?.total_credits ?? 0)}
              description={t('Credits consumed in range')}
              icon={Coins}
              loading={isLoading}
              tone='teal'
            />
            <StatCard
              title={t('Cost estimate')}
              value={`¥${formatNumber(Number((overview?.cost_estimate ?? 0).toFixed(2)))}`}
              description={t('Consumption valued at selling price')}
              icon={TrendingUp}
              loading={isLoading}
              tone='gray'
            />
            <StatCard
              title={t('Total recharge')}
              value={`¥${formatNumber(overview?.total_recharge ?? 0)}`}
              description={t('Successful top-up amount')}
              icon={Wallet}
              loading={isLoading}
              tone='rose'
            />
            <StatCard
              title={t('Users')}
              value={formatNumber(overview?.user_count ?? 0)}
              description={t('{{active}} active in range', {
                active: overview?.active_user_count ?? 0,
              })}
              icon={Users}
              loading={isLoading}
              tone='gray'
            />
          </div>

          <div className='grid gap-6 lg:grid-cols-2'>
            <CreditPoolModelDistribution
              modelCredits={overview?.model_credits ?? {}}
              loading={isLoading}
            />
            <CreditPoolUserStats stats={userStats} />
          </div>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
