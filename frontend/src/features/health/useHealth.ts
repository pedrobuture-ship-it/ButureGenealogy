import { useAsync } from '../../hooks/useAsync'
import { api } from '../../utils/api'

export function useHealth() {
  return useAsync(async () => {
    const [info, db] = await Promise.all([api.getAppInfo(), api.healthCheck()])
    return { info, db }
  })
}
