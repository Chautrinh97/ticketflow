export function formatCurrency(amount: number, currency = 'VND') {
  if (currency === 'VND') {
    return `${new Intl.NumberFormat('vi-VN').format(amount)}đ`
  }
  return new Intl.NumberFormat('vi-VN', { style: 'currency', currency }).format(amount)
}
