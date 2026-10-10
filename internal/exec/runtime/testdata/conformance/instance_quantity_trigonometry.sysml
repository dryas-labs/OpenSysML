// Modified by DRYAS maintainers: explicitly select the SI unit when wildcard imports also expose a same-named function.
// Trigonometric functions over angle quantities: a degree argument is converted
// to radians before the function applies, and a radian argument passes through.
// The SI degree carries the library's inexact factor 1.745329E-02, so results
// of degree arguments are pinned to a tolerance. TrigFunctions::* is imported
// because a bare `sin` resolves only where the model imports it.
package test {
	public import ISQ::*;
	public import SI::*;
	public import QuantityCalculations::*;
	public import TrigFunctions::*;
	private alias rad for SI::radian;

	part def Turn {
		attribute right : AngularMeasureValue = 90 ['°'];
		attribute rise : ScalarValues::Real = sin(right);
		attribute risen : ScalarValues::Boolean = rise > 0.999999 and rise <= 1.0;
		attribute run : ScalarValues::Real = cos(180 ['°']);
		attribute reversed : ScalarValues::Boolean = run < -0.999999 and run >= -1.0;
		attribute flat : ScalarValues::Real = sin(0.0 [rad]);
		attribute half : ScalarValues::Real = sin(30 ['°']);
		attribute halved : ScalarValues::Boolean = half > 0.4999 and half < 0.5001;
		attribute unwound : ScalarValues::Real = arcsin(1.0);
	}
}
