(module (memory 1 1) (data (i32.const 0) "\7f\40\01\c2\83\44\05\c6\87\48\09\ca\8b\4c\0d\ce\8f\50\11\d2\93\54\15\d6\97\58\19\da\9b\5c\1d\de\9f\60\21\e2\a3\64\25\e6\a7\68\29\ea\ab\6c\2d\ee\af\70\31\f2\b3\74\35\f6\b7\78\39\fa\bb\7c\3d\fe\bf\80\41\02\c3\84\45\06\c7\88\49\0a\cb\8c\4d\0e\cf\90\51\12\d3\94\55\16\d7\98\59\1a\db\9c\5d\1e\df\a0\61\22\e3\a4\65\26\e7\a8\69\2a\eb\ac\6d\2e\ef\b0\71\32\f3\b4\75\36\f7\b8\79\3a\fb\bc\7d\3e\ff\c0\81\42\03\c4\85\46\07\c8\89\4a\0b\cc\8d\4e\0f\d0\91\52\13\d4\95\56\17\d8\99\5a\1b\dc\9d\5e\1f\e0\a1\62\23\e4\a5\66\27\e8\a9\6a\2b\ec\ad\6e\2f\f0\b1\72\33\f4\b5\76\37\f8\b9\7a\3b\fc\bd\7e\3f\00\c1\82\43\04\c5\86\47\08\c9\8a\4b\0c\cd\8e\4f\10\d1\92\53\14\d5\96\57\18\d9\9a\5b\1c\dd\9e\5f\20\e1\a2\63\24\e5\a6\67\28\e9\aa\6b\2c\ed\ae\6f\30\f1\b2\73\34\f5\b6\77\38\f9\ba\7b\3c\fd\be")
      (func (export "run") (param i32 i32) (result i32) (local $i i32) (local $v v128) (local $address i32)
        local.get 0 i32x4.splat local.set $v
        (loop $loop local.get $i i32.const 127 i32.and i32.const 1 i32.add local.set $address
                  local.get $address local.get $v v128.store align=1
                  local.get $address local.get 0 local.get $i i32.add i32.store8 offset=16
                  local.get $address v128.load offset=1 align=1
                  v128.const i32x4 1 3 5 7 i32x4.add local.set $v
          local.get $i i32.const 1 i32.add local.tee $i
          local.get 1 i32.lt_u br_if $loop)
        local.get $v i32x4.extract_lane 0 local.get $v i32x4.extract_lane 1 i32.xor local.get $v i32x4.extract_lane 2 i32.xor local.get $v i32x4.extract_lane 3 i32.xor))
