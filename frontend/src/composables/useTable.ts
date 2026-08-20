import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'

export function useTable<T = any>(fetcher: (params: any) => Promise<{ list: T[]; total: number }>, defaultSize = 20) {
  const list = ref<T[]>([]) as any
  const total = ref(0)
  const loading = ref(false)
  const query = reactive<any>({ page: 1, page_size: defaultSize })

  async function load(extra: any = {}) {
    loading.value = true
    try {
      const params = { ...query, ...extra }
      const res = await fetcher(params)
      list.value = res.list || []
      total.value = res.total || 0
    } catch (e: any) {
      list.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  function onPageChange(page: number) {
    query.page = page
    load()
  }
  function onSizeChange(size: number) {
    query.page_size = size
    query.page = 1
    load()
  }
  function search() {
    query.page = 1
    load()
  }

  async function confirm(action: () => Promise<any>, msg = '操作成功') {
    try {
      await action()
      ElMessage.success(msg)
      await load()
    } catch (e: any) {
      // error already toasted by interceptor
    }
  }

  return { list, total, loading, query, load, onPageChange, onSizeChange, search, confirm }
}
