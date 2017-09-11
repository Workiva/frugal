" Vim syntax file
" Language: Frugal
" Maintainer: Tyler Treat <tyler.treat@workiva.com>
" Latest Revision: 1 December 2015

if version < 600
  syntax clear
elseif exists("b:current_syntax")
  finish
endif

" Todo
syn keyword frugalTodo TODO todo FIXME fixme XXX xxx NOTE contained

" Comments
syn match frugalComment "#.*" contains=frugalTodo
syn region frugalComment start="/\*" end="\*/" contains=frugalTodo
syn match frugalComment "//.\{-}\(?>\|$\)\@=" contains=frugalTodo

" String
syn region frugalStringDouble matchgroup=None start=+"+  end=+"+

" Number
syn match frugalNumber "-\=\<\d\+\>" contained

" Keywords
syn keyword frugalKeyword namespace
syn keyword frugalKeyword xsd_all xsd_optional xsd_nillable xsd_attrs
syn keyword frugalKeyword include cpp_include cpp_type const optional required
syn keyword frugalBasicTypes void bool byte i8 i16 i32 i64 double string binary
syn keyword frugalStructure map list set struct typedef exception enum throws union

" Special
syn match frugalSpecial "\d\+:"

" Structure
syn keyword frugalStructure service scope oneway extends prefix

if version >= 508 || !exists("did_frugal_syn_inits")
  if version < 508
    let did_frugal_syn_inits = 1
    command! -nargs=+ HiLink hi link <args>
  else
    command! -nargs=+ HiLink hi def link <args>
  endif

  HiLink frugalComment Comment
  HiLink frugalKeyword Special
  HiLink frugalBasicTypes Type
  HiLink frugalStructure StorageClass
  HiLink frugalTodo Todo
  HiLink frugalString String
  HiLink frugalNumber Number
  HiLink frugalSpecial Special
  HiLink frugalStructure Structure

  delcommand HiLink
endif

let b:current_syntax = "frugal"
