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
import { Input } from '@/components/ui/input'

type InputProps = React.ComponentProps<typeof Input>

export type NumberInputProps = Omit<InputProps, 'value' | 'onChange'> & {
  /** Committed numeric value from form state (or '' when unset). */
  value: number | ''
  /**
   * Called only with finite numbers. The field is allowed to be visibly empty
   * during editing without pushing NaN/empty into form state — the previous
   * committed value is preserved until the user types a valid number again.
   */
  onValueChange: (value: number) => void
}

/**
 * Number input that lets the user fully clear the field and retype, while never
 * writing NaN into form state.
 *
 * The native `<input type="number">` controlled by form state snaps back to the
 * last value when emptied (because empty -> NaN is ignored). We fix this by
 * showing a local string "draft" only while the field is focused (it may be
 * empty), and showing the committed value otherwise. A committed number is only
 * pushed up when the draft parses to a finite number. All state updates happen
 * in event handlers — no effects, no render-time mutations.
 */
export function NumberInput({ value, onValueChange, ...props }: NumberInputProps) {
  const committed = value === '' ? '' : String(value)
  const [draft, setDraft] = React.useState<string>('')
  const [focused, setFocused] = React.useState(false)

  // While focused, the local draft drives the field (can be empty / partial).
  // While not focused, the committed value from form state drives it, so
  // external resets/async loads are reflected immediately.
  const display = focused ? draft : committed

  return (
    <Input
      {...props}
      type='number'
      value={display}
      onFocus={(e) => {
        setDraft(committed)
        setFocused(true)
        props.onFocus?.(e)
      }}
      onBlur={(e) => {
        setFocused(false)
        props.onBlur?.(e)
      }}
      onChange={(e) => {
        const raw = e.target.value
        setDraft(raw)
        if (raw === '') return // allow empty while editing; do not commit
        const next = Number(raw)
        if (Number.isFinite(next)) onValueChange(next)
      }}
    />
  )
}
