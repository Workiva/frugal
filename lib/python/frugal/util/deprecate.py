import functools
import warnings

import six


def deprecated(func):
    '''This is a decorator which can be used to mark functions
    as deprecated. It will result in a warning being emitted
    when the function is used.'''

    @functools.wraps(func)
    def new_func(*args, **kwargs):
        func_code = six.get_function_code(func)
        warnings.warn_explicit(
            "Call to deprecated function {}.".format(func.__name__),
            category=DeprecationWarning,
            filename=func_code.co_filename,
            lineno=func_code.co_firstlineno + 1
        )
        return func(*args, **kwargs)
    return new_func
