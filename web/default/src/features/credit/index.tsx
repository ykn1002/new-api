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
import { useQuery } from '@tanstack/react-query'
import { AlertTriangle, Coins } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatQuota, formatTimestampToDate } from '@/lib/format'
import { SectionPageLayout } from '@/components/layout'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import { getCreditBatches } from './api'
import { CreditSourceCards } from './components/credit-source-cards'

export function Credit() {
  const { t } = useTranslation()

  const { data, isLoading } = useQuery({
    queryKey: ['credit-batches'],
    queryFn: getCreditBatches,
  })

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Credit Composition')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='mx-auto max-w-4xl space-y-6'>
          {/* 总余额 */}
          <div className='rounded-lg border p-5'>
            <div className='flex items-center gap-2'>
              <Coins className='text-muted-foreground/60 size-4' />
              <span className='text-muted-foreground text-xs font-medium tracking-wider uppercase'>
                {t('Total credits')}
              </span>
            </div>
            {isLoading ? (
              <Skeleton className='mt-2 h-9 w-40' />
            ) : (
              <div className='text-foreground mt-2 font-mono text-3xl font-bold tabular-nums'>
                {formatQuota(data?.total ?? 0)}
              </div>
            )}
          </div>

          {/* 即将过期提示 */}
          {data?.expiring_soon && data.expiring_soon.remaining > 0 && (
            <Alert>
              <AlertTriangle className='size-4' />
              <AlertTitle>{t('Credits expiring soon')}</AlertTitle>
              <AlertDescription>
                {t(
                  '{{amount}} credits will expire on {{date}}',
                  {
                    amount: formatQuota(data.expiring_soon.remaining),
                    date: formatTimestampToDate(data.expiring_soon.expire_at),
                  }
                )}
              </AlertDescription>
            </Alert>
          )}

          {/* 来源构成 */}
          <CreditSourceCards
            sources={data?.sources ?? []}
            total={data?.total ?? 0}
            loading={isLoading}
          />
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
