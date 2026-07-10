# VRM をここに置く
#
# デモでは 2 体分が必要:
#   self.vrm   … プレイヤー（自分）のアバター  → personaKind: "self"
#   avatar.vrm … 会話相手（保有ペルソナ）の VRM → personaKind: "owned"
#
# 同じ VRM で試す場合:
#   cp your-model.vrm self.vrm
#   cp your-model.vrm avatar.vrm
#
# パスは src/lib/demoConfig.ts または .env の NEXT_PUBLIC_SELF_VRM_URL / NEXT_PUBLIC_VRM_URL で変更可
